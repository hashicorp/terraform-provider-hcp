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

func TestAccVaultSecretsResourceIntegrationGitLab(t *testing.T) {
	accessToken := checkRequiredEnvVarOrFail(t, "GITLAB_ACCESS_TOKEN")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access token
			{
				Config: gitlabConfig(integrationName1, accessToken),
				Check: resource.ComposeTestCheckFunc(
					gitlabCheckFuncs(integrationName1, accessToken)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: gitlabConfig(integrationName2, accessToken),
				Check: resource.ComposeTestCheckFunc(
					gitlabCheckFuncs(integrationName2, accessToken)...,
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.DeleteIntegration(&secret_service.DeleteIntegrationParams{
						Name:           integrationName2,
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: gitlabConfig(integrationName2, accessToken),
				Check: resource.ComposeTestCheckFunc(
					gitlabCheckFuncs(integrationName2, accessToken)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.CreateIntegration(&secret_service.CreateIntegrationParams{
						Body: &secretmodels.SecretServiceCreateIntegrationBody{
							Name:     integrationName2,
							Provider: "gitlab",
							Capabilities: []*secretmodels.Secrets20231128Capability{
								secretmodels.Secrets20231128CapabilitySYNC.Pointer(),
							},
							GitlabAccessToken: &secretmodels.Secrets20231128GitlabAccessTokenRequest{
								Token: accessToken,
							},
						},
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: gitlabConfig(integrationName2, accessToken),
				Check: resource.ComposeTestCheckFunc(
					gitlabCheckFuncs(integrationName2, accessToken)...,
				),
				ResourceName:  "hcp_vault_secrets_integration.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if integrationExists(t, integrationName1) {
				return fmt.Errorf("test GitLab integration %s was not destroyed", integrationName1)
			}
			if integrationExists(t, integrationName2) {
				return fmt.Errorf("test GitLab integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func gitlabCheckFuncs(integrationName, accessToken string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.0", "SYNC"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "provider_type", "gitlab"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "gitlab_access.token", accessToken),
	}
}

func gitlabConfig(integrationName, accessToken string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration" "acc_test" {
		name = %q
		capabilities = ["SYNC"]
		provider_type = "gitlab"
		gitlab_access = {
			token = %q
		}
    }`, integrationName, accessToken)
}

func integrationExists(t *testing.T, name string) bool {
	t.Helper()
	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetIntegration(
		secret_service.NewGetIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)

	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
