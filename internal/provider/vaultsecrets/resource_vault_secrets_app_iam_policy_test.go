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

func TestAccVaultSecretsAppIamPolicyResource(t *testing.T) {
	appName := generateRandomSlug()
	projectID := os.Getenv("HCP_PROJECT_ID")
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/secrets.app-manager"
	roleName2 := "roles/secrets.app-secret-reader"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVaultSecretsAppIamPolicy(projectName, roleName, appName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_vault_secrets_app_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
			{
				Config: testAccVaultSecretsAppIamPolicy(projectName, roleName2, appName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_vault_secrets_app_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccVaultSecretsAppIamBindingResource(t *testing.T) {
	appName := generateRandomSlug()
	projectID := os.Getenv("HCP_PROJECT_ID")
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/secrets.app-secret-reader"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVaultSecretsAppIamBinding(projectName, appName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccVaultSecretsAppIamPolicy(projectName, roleName, appName string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = "test-sp"
	parent = %q
}

data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = %q
      principals = [
		hcp_service_principal.example.resource_id,
      ]
    },
  ]
}

resource "hcp_vault_secrets_app" "example" {
	app_name    = %q
	description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_policy" "example" {
    resource_name = hcp_vault_secrets_app.example.resource_name
    policy_data = data.hcp_iam_policy.example.policy_data
}
`, projectName, roleName, appName)
}

func testAccVaultSecretsAppIamBinding(projectName, appName, roleName string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = "hvs-sp"
	parent = %q
}

resource "hcp_vault_secrets_app" "example" {
	app_name    = %q
	description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_binding" "example" {
	resource_name = hcp_vault_secrets_app.example.resource_name
	principal_id = hcp_service_principal.example.resource_id
	role         = %q
}
`, projectName, appName, roleName)
}
