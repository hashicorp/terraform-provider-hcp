// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceSecret(t *testing.T) {
	testAppName1 := generateRandomSlug()
	testAppName2 := generateRandomSlug()
	resource.ParallelTest(t, resource.TestCase{
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
				Config: fmt.Sprintf(`
				resource "hcp_project" "example" {
					name        = "test-project"
				}

				resource "hcp_vault_secrets_app" "example" {
					app_name = %q
					description = "Acceptance test run"
					project_id = hcp_project.example.resource_id
			  	}

				resource "hcp_vault_secrets_secret" "example" {
					app_name = hcp_vault_secrets_app.example.app_name
					secret_name = "test_secret"
					secret_value = "super secret"
					project_id = hcp_project.example.resource_id
				}`, testAppName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", "test_secret"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_secret.example", "project_id"),
				),
			},
		},
	})
}
