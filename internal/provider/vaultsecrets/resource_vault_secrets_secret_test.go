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
	testAppName := generateRandomSlug()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, testAppName)
				},
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = "test_secret"
					secret_value = "super secret"
				}`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", "a_long_and_complicated_secret_name_but_less_than_64_characters"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
				),
			},
		},
	})
}
