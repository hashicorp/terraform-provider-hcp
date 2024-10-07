// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceIntegrationMongoDBAtlas(t *testing.T) {
	publicKey := checkRequiredEnvVarOrFail(t, "MONGODBATLAS_API_PUBLIC_KEY")
	privateKey := checkRequiredEnvVarOrFail(t, "MONGODBATLAS_API_PRIVATE_KEY")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: mongoDBAtlasConfig(integrationName1, publicKey, privateKey),
				Check: resource.ComposeTestCheckFunc(
					mongoDBAtlasCheckFuncs(integrationName1, publicKey, privateKey)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: mongoDBAtlasConfig(integrationName2, publicKey, privateKey),
				Check: resource.ComposeTestCheckFunc(
					mongoDBAtlasCheckFuncs(integrationName2, publicKey, privateKey)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: mongoDBAtlasConfig(integrationName2, publicKey, privateKey),
				Check: resource.ComposeTestCheckFunc(
					mongoDBAtlasCheckFuncs(integrationName2, publicKey, privateKey)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecretsPreview.DeleteMongoDBAtlasIntegration(&secret_service.DeleteMongoDBAtlasIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: mongoDBAtlasConfig(integrationName2, publicKey, privateKey),
				Check: resource.ComposeTestCheckFunc(
					mongoDBAtlasCheckFuncs(integrationName2, publicKey, privateKey)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecretsPreview.CreateMongoDBAtlasIntegration(&secret_service.CreateMongoDBAtlasIntegrationParams{
						Body: &secretmodels.SecretServiceCreateMongoDBAtlasIntegrationBody{
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							StaticCredentialDetails: &secretmodels.Secrets20231128MongoDBAtlasStaticCredentialsRequest{
								APIPublicKey:  publicKey,
								APIPrivateKey: privateKey,
							},
							Name: integrationName2,
						},
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: mongoDBAtlasConfig(integrationName2, publicKey, privateKey),
				Check: resource.ComposeTestCheckFunc(
					mongoDBAtlasCheckFuncs(integrationName2, publicKey, privateKey)...,
				),
				ResourceName:  "hcp_vault_secrets_integration_mongodbatlas.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if mongoDBAtlasIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test mongo db atlas integration %s was not destroyed", integrationName1)
			}
			if mongoDBAtlasIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test mongo db atlas integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func mongoDBAtlasConfig(integrationName, apiPublicKey, apiPrivateKey string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration_mongodbatlas" "acc_test"{
        name = %q
        capabilities = ["ROTATION"]
        static_credential_details = {
          api_public_key = %q
          api_private_key = %q
        }
    }`, integrationName, apiPublicKey, apiPrivateKey)
}

func mongoDBAtlasCheckFuncs(integrationName, apiPublicKey, apiPrivateKey string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_mongodbatlas.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_mongodbatlas.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_mongodbatlas.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "static_credential_details.api_public_key", apiPublicKey),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_mongodbatlas.acc_test", "static_credential_details.api_private_key", apiPrivateKey),
	}
}

func mongoDBAtlasIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetMongoDBAtlasIntegration(
		secret_service.NewGetMongoDBAtlasIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
