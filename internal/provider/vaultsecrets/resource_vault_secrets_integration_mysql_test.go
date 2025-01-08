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

func TestAccVaultSecretsResourceIntegrationMysql(t *testing.T) {
	connectionString := checkRequiredEnvVarOrFail(t, "MYSQL_CONNECTION_STRING")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: mysqlConfig(integrationName1, connectionString),
				Check: resource.ComposeTestCheckFunc(
					mysqlCheckFuncs(integrationName1, connectionString)...,
				),
			},
			// Changing the name forces a recreation
			{
				Config: mysqlConfig(integrationName2, connectionString),
				Check: resource.ComposeTestCheckFunc(
					mysqlCheckFuncs(integrationName2, connectionString)...,
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: mysqlConfig(integrationName2, connectionString),
				Check: resource.ComposeTestCheckFunc(
					mysqlCheckFuncs(integrationName2, connectionString)...,
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
				Config: mysqlConfig(integrationName2, connectionString),
				Check: resource.ComposeTestCheckFunc(
					mysqlCheckFuncs(integrationName2, connectionString)...,
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
							Capabilities: []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityROTATION.Pointer()},
							MysqlStaticCredentials: &secretmodels.Secrets20231128MysqlStaticCredentialsRequest{
								ConnectionString: connectionString,
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
				Config: mysqlConfig(integrationName2, connectionString),
				Check: resource.ComposeTestCheckFunc(
					mysqlCheckFuncs(integrationName2, connectionString)...,
				),
				ResourceName:  "hcp_vault_secrets_integration.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if mysqlIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test mysql integration %s was not destroyed", integrationName1)
			}
			if mysqlIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test mysql integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func mysqlConfig(integrationName, connectionString string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_integration" "acc_test" {
		name = %q
		capabilities = ["ROTATION"]
        provider_type = "mysql"
		mysql_static_credentials = {
			connection_string = %q
	   }
    }`, integrationName, connectionString)
}

func mysqlCheckFuncs(integrationName, connectionString string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.#", "1"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "capabilities.0", "ROTATION"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "provider_type", "mysql"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "mysql_static_credentials.connection_string", connectionString),
	}
}

func mysqlIntegrationExists(t *testing.T, name string) bool {
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
