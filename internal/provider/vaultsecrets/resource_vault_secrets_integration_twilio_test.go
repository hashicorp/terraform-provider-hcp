// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceIntegrationTwilio(t *testing.T) {
	accountSID := checkRequiredEnvVarOrFail(t, "TWILIO_ACCOUNT_SID")
	apiKeySID := checkRequiredEnvVarOrFail(t, "TWILIO_API_KEY_SID")
	apiKeySecret := checkRequiredEnvVarOrFail(t, "TWILIO_API_KEY_SECRET")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: twilioConfig(integrationName1, accountSID, apiKeySID, apiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					twilioCheckFuncs(integrationName1, accountSID, apiKeySID, apiKeySecret)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: twilioConfig(integrationName2, accountSID, apiKeySID, apiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					twilioCheckFuncs(integrationName2, accountSID, apiKeySID, apiKeySecret)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: twilioConfig(integrationName2, accountSID, apiKeySID, apiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					twilioCheckFuncs(integrationName2, accountSID, apiKeySID, apiKeySecret)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.DeleteTwilioIntegration(&secret_service.DeleteTwilioIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: twilioConfig(integrationName2, accountSID, apiKeySID, apiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					twilioCheckFuncs(integrationName2, accountSID, apiKeySID, apiKeySecret)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.CreateTwilioIntegration(&secret_service.CreateTwilioIntegrationParams{
						Body: &secretmodels.SecretServiceCreateTwilioIntegrationBody{
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							StaticCredentialDetails: &secretmodels.Secrets20231128TwilioStaticCredentialsRequest{
								AccountSid:   accountSID,
								APIKeySid:    apiKeySID,
								APIKeySecret: apiKeySecret,
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
				Config: twilioConfig(integrationName2, accountSID, apiKeySID, apiKeySecret),
				Check: resource.ComposeTestCheckFunc(
					twilioCheckFuncs(integrationName2, accountSID, apiKeySID, apiKeySecret)...,
				),
				ResourceName:  "hcp_vault_secrets_integration.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if twilioIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test twilio integration %s was not destroyed", integrationName1)
			}
			if twilioIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test twilio integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func twilioConfig(integrationName, accountSID, apiKeySID, apiKeySecret string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
        provider_type = "twilio"
		twilio_static_credentials = {
			account_sid = %q
			api_key_sid = %q
			api_key_secret = %q
	   }
    }`, integrationName, accountSID, apiKeySID, apiKeySecret)
}

func twilioCheckFuncs(integrationName, accountSID, apiKeySID, apiKeySecret string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "provider_type", "twilio"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "twilio_static_credentials.account_sid", accountSID),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "twilio_static_credentials.api_key_secret", apiKeySecret),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "twilio_static_credentials.api_key_sid", apiKeySID),
	}
}

func twilioIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetTwilioIntegration(
		secret_service.NewGetTwilioIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
