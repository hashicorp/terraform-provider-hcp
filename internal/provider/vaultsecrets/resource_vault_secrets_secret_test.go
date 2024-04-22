// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceSecret(t *testing.T) {
	testAppName1 := generateRandomSlug()
	testAppName2 := generateRandomSlug()
	projectID := os.Getenv("HCP_PROJECT_ID")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{

			{
				PreConfig: func() {
					createTestApp(t, testAppName1)
				},
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = "test_secret"
					secret_value = "super secret"
				}`, testAppName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName1),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", "test_secret"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
				),
			},
			{
				PreConfig: func() {
					createTestApp(t, testAppName2)
				},
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = "test_secret"
					secret_value = "super secret"
				}`, testAppName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", "test_secret"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "project_id", projectID),
				),
			},
		},
	})
}
