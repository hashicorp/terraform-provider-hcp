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

func TestAccVaultSecretsResourceIntegrationAzure(t *testing.T) {
	tenantID := checkRequiredEnvVarOrFail(t, "AZURE_TENANT_ID")
	clientID := checkRequiredEnvVarOrFail(t, "AZURE_CLIENT_ID")
	clientSecret := checkRequiredEnvVarOrFail(t, "AZURE_CLIENT_SECRET")
	audience := checkRequiredEnvVarOrFail(t, "AZURE_INTEGRATION_AUDIENCE")

	integrationName1 := generateRandomSlug()
	// Set the integration name that is configured in the subject claim while creating a federated credential.
	integrationName2 := checkRequiredEnvVarOrFail(t, "AZURE_INTEGRATION_NAME_WIF")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: azureClientSecretConfig(integrationName1, clientID, tenantID, clientSecret),
				Check: resource.ComposeTestCheckFunc(
					azureCheckClientSecretKeyFuncs(integrationName1, clientID, tenantID, clientSecret)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: azureClientSecretConfig(integrationName2, clientID, tenantID, clientSecret),
				Check: resource.ComposeTestCheckFunc(
					azureCheckClientSecretKeyFuncs(integrationName2, clientID, tenantID, clientSecret)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: azureFederatedIdentityConfig(integrationName2, clientID, tenantID, audience),
				Check: resource.ComposeTestCheckFunc(
					azureCheckFederatedIdentityFuncs(integrationName2, clientID, tenantID, audience)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.DeleteAzureIntegration(&secret_service.DeleteAzureIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: azureFederatedIdentityConfig(integrationName2, clientID, tenantID, audience),
				Check: resource.ComposeTestCheckFunc(
					azureCheckFederatedIdentityFuncs(integrationName2, clientID, tenantID, audience)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.CreateAzureIntegration(&secret_service.CreateAzureIntegrationParams{
						Body: &secretmodels.SecretServiceCreateAzureIntegrationBody{
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							FederatedWorkloadIdentity: &secretmodels.Secrets20231128AzureFederatedWorkloadIdentityRequest{
								Audience: audience,
								ClientID: clientID,
								TenantID: tenantID,
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
				Config: azureFederatedIdentityConfig(integrationName2, clientID, tenantID, audience),
				Check: resource.ComposeTestCheckFunc(
					azureCheckFederatedIdentityFuncs(integrationName2, clientID, tenantID, audience)...,
				),
				ResourceName:  "hcp_vault_secrets_integration.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
				PlanOnly:      true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if azureIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test azure integration %s was not destroyed", integrationName1)
			}
			if azureIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test azure integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func azureClientSecretConfig(integrationName, clientID, tenantID, clientSecret string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
        provider_type = "azure"
		azure_client_secret = {
			tenant_id = %q
			client_id = %q
			client_secret = %q
	   }
    }`, integrationName, tenantID, clientID, clientSecret)
}

func azureFederatedIdentityConfig(integrationName, clientID, tenantID, audience string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
         provider_type = "azure"
		azure_federated_workload_identity = {
			tenant_id = %q
			client_id = %q
			audience = %q
	   }
    }`, integrationName, tenantID, clientID, audience)
}

func azureCheckClientSecretKeyFuncs(integrationName, clientID, tenantID, clientSecret string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "provider_type", "azure"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_client_secret.client_id", clientID),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_client_secret.tenant_id", tenantID),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_client_secret.client_secret", clientSecret),
	}
}

func azureCheckFederatedIdentityFuncs(integrationName, clientID, tenantID, audience string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "provider_type", "azure"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_federated_workload_identity.audience", audience),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_federated_workload_identity.tenant_id", tenantID),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "azure_federated_workload_identity.client_id", clientID),
	}
}

func azureIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetAzureIntegration(
		secret_service.NewGetAzureIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
