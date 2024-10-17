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

func TestAccVaultSecretsResourceIntegrationConfluent(t *testing.T) {
	cloudApiKeyID := checkRequiredEnvVarOrFail(t, "CONFLUENT_API_KEY_ID")
	cloudApiKeySecret := checkRequiredEnvVarOrFail(t, "CONFLUENT_API_SECRET")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: confluentConfig(integrationName1, cloudApiKeyID, cloudApiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					confluentCheckFuncs(integrationName1, cloudApiKeyID, cloudApiKeySecret)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: confluentConfig(integrationName2, cloudApiKeyID, cloudApiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					confluentCheckFuncs(integrationName2, cloudApiKeyID, cloudApiKeySecret)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: confluentConfig(integrationName2, cloudApiKeyID, cloudApiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					confluentCheckFuncs(integrationName2, cloudApiKeyID, cloudApiKeySecret)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecretsPreview.DeleteConfluentIntegration(&secret_service.DeleteConfluentIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: confluentConfig(integrationName2, cloudApiKeyID, cloudApiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					confluentCheckFuncs(integrationName2, cloudApiKeyID, cloudApiKeySecret)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecretsPreview.CreateConfluentIntegration(&secret_service.CreateConfluentIntegrationParams{
						Body: &secretmodels.SecretServiceCreateConfluentIntegrationBody{
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							StaticCredentialDetails: &secretmodels.Secrets20231128ConfluentStaticCredentialsRequest{
								CloudAPIKeyID:  cloudApiKeyID,
								CloudAPISecret: cloudApiKeySecret,
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
				Config: confluentConfig(integrationName2, cloudApiKeyID, cloudApiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					confluentCheckFuncs(integrationName2, cloudApiKeyID, cloudApiKeySecret)...,
				),
				ResourceName:  "hcp_vault_secrets_integration_confluent.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if confluentIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test confluent integration %s was not destroyed", integrationName1)
			}
			if confluentIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test confluent integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func confluentConfig(integrationName, cloudApiKeyID, cloudApiSecret string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration_confluent" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
		static_credential_details = {
			cloud_api_key_id = %q
			cloud_api_secret = %q
	   }
    }`, integrationName, cloudApiKeyID, cloudApiSecret)
}

func confluentCheckFuncs(integrationName, cloudApiKeyID, cloudApiKeySecret string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_confluent.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_confluent.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_confluent.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "static_credential_details.cloud_api_secret", cloudApiKeySecret),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_confluent.acc_test", "static_credential_details.cloud_api_key_id", cloudApiKeyID),
	}
}

func confluentIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetConfluentIntegration(
		secret_service.NewGetConfluentIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
