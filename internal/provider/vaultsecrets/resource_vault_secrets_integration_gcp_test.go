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

func TestAccVaultSecretsResourceIntegrationGCP(t *testing.T) {
	projectID := checkRequiredEnvVarOrFail(t, "GCP_PROJECT_ID")
	clientEmail := checkRequiredEnvVarOrFail(t, "GCP_CLIENT_EMAIL")
	serviceAccountKey := checkRequiredEnvVarOrFail(t, "GCP_SERVICE_ACCOUNT_KEY")
	serviceAccountEmail := checkRequiredEnvVarOrFail(t, "GCP_INTEGRATION_SERVICE_ACCOUNT_EMAIL")
	audience := checkRequiredEnvVarOrFail(t, "GCP_INTEGRATION_AUDIENCE")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: gcpServiceAccountConfig(integrationName1, serviceAccountKey),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckServiceAccountKeyFuncs(integrationName1, clientEmail, serviceAccountKey, projectID)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: gcpServiceAccountConfig(integrationName2, serviceAccountKey),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckServiceAccountKeyFuncs(integrationName2, clientEmail, serviceAccountKey, projectID)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: gcpFederatedIdentityConfig(integrationName2, serviceAccountEmail, audience),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFederatedIdentityFuncs(integrationName2, audience, serviceAccountEmail)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.DeleteGcpIntegration(&secret_service.DeleteGcpIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: gcpFederatedIdentityConfig(integrationName2, serviceAccountEmail, audience),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFederatedIdentityFuncs(integrationName2, audience, serviceAccountEmail)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.CreateGcpIntegration(&secret_service.CreateGcpIntegrationParams{
						Body: &secretmodels.SecretServiceCreateGcpIntegrationBody{
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							FederatedWorkloadIdentity: &secretmodels.Secrets20231128GcpFederatedWorkloadIdentityRequest{
								Audience:            audience,
								ServiceAccountEmail: serviceAccountEmail,
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
				Config: gcpFederatedIdentityConfig(integrationName2, serviceAccountEmail, audience),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFederatedIdentityFuncs(integrationName2, audience, serviceAccountEmail)...,
				),
				ResourceName:  "hcp_vault_secrets_integration_gcp.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if gcpIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test gcp integration %s was not destroyed", integrationName1)
			}
			if gcpIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test gcp integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func gcpServiceAccountConfig(integrationName, serviceAccountKey string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration_gcp" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
		service_account_key = {
			credentials = %q
	   }
    }`, integrationName, serviceAccountKey)
}

func gcpFederatedIdentityConfig(integrationName, serviceAccountEmail, audience string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration_gcp" "acc_test" {
		name = %q
		capabilities = ["DYNAMIC"]
		federated_workload_identity = {
			service_account_email = %q
			audience = %q
	   }
    }`, integrationName, serviceAccountEmail, audience)
}

func gcpCheckServiceAccountKeyFuncs(integrationName, clientEmail, credentials, projectID string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "service_account_key.client_email", clientEmail),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "service_account_key.credentials", credentials),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "service_account_key.project_id", projectID),
	}
}

func gcpCheckFederatedIdentityFuncs(integrationName, audience, serviceAccountEmail string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_gcp.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "capabilities.0", "DYNAMIC"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "federated_workload_identity.audience", audience),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test", "federated_workload_identity.service_account_email", serviceAccountEmail),
	}
}

func gcpIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetGcpIntegration(
		secret_service.NewGetGcpIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
