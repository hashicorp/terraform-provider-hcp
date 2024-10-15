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

func TestAccVaultSecretsResourceIntegrationAWS(t *testing.T) {
	accessKeyID := checkRequiredEnvVarOrFail(t, "AWS_ACCESS_KEY_ID")
	secretAccessKey := checkRequiredEnvVarOrFail(t, "AWS_SECRET_ACCESS_KEY")
	roleArn := checkRequiredEnvVarOrFail(t, "AWS_INTEGRATION_ROLE_ARN")
	audience := checkRequiredEnvVarOrFail(t, "AWS_INTEGRATION_AUDIENCE")

	integrationName1 := generateRandomSlug()
	integrationName2 := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial integration with access keys
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration_aws" "acc_test" {
						name = %q
						capabilities = ["DYNAMIC", "ROTATION"]
						access_keys = {
			               access_key_id = %q
			               secret_access_key = %q
			           }
				  }`, integrationName1, accessKeyID, secretAccessKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_name"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "name", integrationName1),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.#", "2"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.0", "DYNAMIC"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.1", "ROTATION"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "access_keys.access_key_id", accessKeyID),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "access_keys.secret_access_key", secretAccessKey),
				),
			},
			// Changing the name forces a recreation
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration_aws" "acc_test" {
						name = %q
						capabilities = ["DYNAMIC", "ROTATION"]
						access_keys = {
			               access_key_id = %q
			               secret_access_key = %q
			           }
				  }`, integrationName2, accessKeyID, secretAccessKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_name"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "name", integrationName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.#", "2"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.0", "DYNAMIC"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.1", "ROTATION"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "access_keys.access_key_id", accessKeyID),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "access_keys.secret_access_key", secretAccessKey),
				),
			},
			// Modifying mutable fields causes an update
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration_aws" "acc_test" {
						name = %q
						capabilities = ["DYNAMIC"]
						federated_workload_identity = {
			               role_arn = %q
			               audience = %q
			           }
				  }`, integrationName2, roleArn, audience),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_name"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "name", integrationName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.0", "DYNAMIC"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.role_arn", roleArn),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.audience", audience),
				),
			},
			// Deleting the integration out of band causes a recreation
			{
				PreConfig: func() {
					deleteTestAwsIntegration(t, integrationName2)
				},
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration_aws" "acc_test" {
						name = %q
						capabilities = ["DYNAMIC"]
						federated_workload_identity = {
                            role_arn = %q
                            audience = %q
                        }
				  }`, integrationName2, roleArn, audience),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_name"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "name", integrationName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.0", "DYNAMIC"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.role_arn", roleArn),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.audience", audience),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing integration can be imported
			{
				PreConfig: func() {
					createTestAwsIntegration(t, integrationName2, roleArn, audience, []*secretmodels.Secrets20231128Capability{secretmodels.Secrets20231128CapabilityDYNAMIC.Pointer()})
				},
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration_aws" "acc_test" {
						name = %q
						capabilities = ["DYNAMIC"]
						federated_workload_identity = {
                            role_arn = %q
                            audience = %q
                        }
				  }`, integrationName2, roleArn, audience),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration_aws.acc_test", "resource_name"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "name", integrationName2),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "capabilities.0", "DYNAMIC"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.role_arn", roleArn),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration_aws.acc_test", "federated_workload_identity.audience", audience),
				),
				ResourceName:  "hcp_vault_secrets_integration_aws.acc_test",
				ImportStateId: integrationName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if awsIntegrationExists(t, integrationName1) {
				return fmt.Errorf("test aws integration %s was not destroyed", integrationName1)
			}
			if awsIntegrationExists(t, integrationName2) {
				return fmt.Errorf("test aws integration %s was not destroyed", integrationName2)
			}
			return nil
		},
	})
}

func awsIntegrationExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetAwsIntegration(
		secret_service.NewGetAwsIntegrationParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Integration != nil
}
