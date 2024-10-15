// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceRotatingSecret(t *testing.T) {
	if _, exists := os.LookupEnv("AWS_ROTATING_SECRET_ACC_ENABLED"); exists {
		testAccVaultSecretsResourceRotatingSecretAWS(t)
	}
}

func testAccVaultSecretsResourceRotatingSecretAWS(t *testing.T) {
	username := checkRequiredEnvVarOrFail(t, "HVS_ROTATING_SECRET_USERNAME")
	integrationName := checkRequiredEnvVarOrFail(t, "HVS_ROTATING_SECRET_INTEGRATION_NAME")
	appName := checkRequiredEnvVarOrFail(t, "HVS_APP_NAME")
	rotationPolicy := "built-in:60-days-2-active"
	secretName1 := "acc_tests_aws_1"
	secretName2 := "acc_tests_aws_2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial rotating secret
			{
				Config: awsRotatingSecretConfig(appName, secretName1, integrationName, rotationPolicy, username),
				Check: resource.ComposeTestCheckFunc(
					awsRotationCheckFunc(appName, secretName1, integrationName, rotationPolicy, username)...,
				),
			},
			// Changing an immutable field causes a recreation
			{
				Config: awsRotatingSecretConfig(appName, secretName2, integrationName, rotationPolicy, username),
				Check: resource.ComposeTestCheckFunc(
					awsRotationCheckFunc(appName, secretName2, integrationName, rotationPolicy, username)...,
				),
			},
			// Deleting the secret out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecretsPreview.DeleteAppSecret(&secret_service.DeleteAppSecretParams{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
						AppName:        appName,
						SecretName:     secretName2,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: awsRotatingSecretConfig(appName, secretName2, integrationName, rotationPolicy, username),
				Check: resource.ComposeTestCheckFunc(
					awsRotationCheckFunc(appName, secretName2, integrationName, rotationPolicy, username)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if awsRotatingSecretExists(t, appName, secretName1) {
				return fmt.Errorf("test rotating secret %s was not destroyed", secretName1)
			}
			if awsRotatingSecretExists(t, appName, secretName2) {
				return fmt.Errorf("test rotating secret %s was not destroyed", secretName2)
			}
			return nil
		},
	})
}

func awsRotatingSecretConfig(appName, name, integrationName, policy, iamUsername string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_rotating_secret" "acc_test_aws" {
	  app_name             = %q
	  secret_provider      = "aws"
	  name                 = %q
	  integration_name     = %q
	  rotation_policy_name = %q
	  aws_access_keys = {
		iam_username = %q
	  }
	}`, appName, name, integrationName, policy, iamUsername)
}

func awsRotationCheckFunc(appName, name, integrationName, policy, iamUsername string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_rotating_secret.acc_test_aws", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "app_name", appName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "secret_provider", "aws"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "name", name),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "integration_name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "rotation_policy_name", policy),
		resource.TestCheckResourceAttr("hcp_vault_secrets_rotating_secret.acc_test_aws", "aws_access_keys.iam_username", iamUsername),
	}
}

func awsRotatingSecretExists(t *testing.T, appName, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetAwsIAMUserAccessKeyRotatingSecretConfig(
		secret_service.NewGetAwsIAMUserAccessKeyRotatingSecretConfigParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithAppName(appName).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Config != nil
}
