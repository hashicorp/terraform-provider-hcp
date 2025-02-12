// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceSecret(t *testing.T) {
	testAppName1 := generateRandomSlug()
	testAppName2 := generateRandomSlug()
	secretName1 := "acc_tests_secret_1"
	secretName2 := "acc_tests_secret_2"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create secretName1 in testAppName1
			{
				PreConfig: func() {
					createTestApp(t, testAppName1)
				},
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = %q
					secret_value = "super secret"
				}`, testAppName1, secretName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName1),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", secretName1),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
				),
			},
			// Changing secret name should cause recreation.
			// Validate that secretName2 is created and secretName1 is destroyed.
			{
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = %q
					secret_value = "super secret"
				}`, testAppName1, secretName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName1),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", secretName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
					testAccCheckSecretExists(t, testAppName1, secretName1),
				),
			},
			// Changing app name should also cause secret recreation.
			// Validate secretName2 is created in testAppName2 and is destroyed in testAppName1
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
					secret_name = %q
					secret_value = "super secret"
					project_id = hcp_project.example.resource_id
				}`, testAppName2, secretName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", secretName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_secret.example", "project_id"),
					testAccCheckSecretExists(t, testAppName1, secretName2),
				),
			},
		},
	})
}

func testAccCheckSecretExists(t *testing.T, appName string, secretName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if secretExists(t, appName, secretName) {
			return fmt.Errorf("test secret %s was not destroyed", secretName)
		}
		return nil
	}
}

func secretExists(t *testing.T, appName string, secretName string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetAppSecret(
		secret_service.NewGetAppSecretParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithAppName(appName).
			WithSecretName(secretName), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Secret != nil
}
