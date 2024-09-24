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

func TestAccVaultSecretsResourceDynamicSecret(t *testing.T) {
	if _, exists := os.LookupEnv("AWS_DYNAMIC_SECRET_ACC_ENABLED"); exists {
		testAccVaultSecretsResourceDynamicSecretAWS(t)
	}

	if _, exists := os.LookupEnv("GCP_DYNAMIC_SECRET_ACC_ENABLED"); exists {
		testAccVaultSecretsResourceDynamicSecretGCP(t)
	}
}

func testAccVaultSecretsResourceDynamicSecretAWS(t *testing.T) {
	roleARN := checkRequiredEnvVarOrFail(t, "HVS_DYNAMIC_SECRET_ROLE_ARN")
	integrationName := checkRequiredEnvVarOrFail(t, "HVS_DYNAMIC_SECRET_INTEGRATION_NAME")
	appName := checkRequiredEnvVarOrFail(t, "HVS_APP_NAME")
	ttl1 := "901s"
	ttl2 := "902s"
	secretName1 := "acc_tests_aws_1"
	secretName2 := "acc_tests_aws_2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial dynamic secret
			{
				Config: awsDynamicSecretConfig(appName, secretName1, integrationName, ttl1, roleARN),
				Check: resource.ComposeTestCheckFunc(
					awsCheckFunc(appName, secretName1, integrationName, ttl1, roleARN)...,
				),
			},
			// Changing an immutable field causes a recreation
			{
				Config: awsDynamicSecretConfig(appName, secretName2, integrationName, ttl1, roleARN),
				Check: resource.ComposeTestCheckFunc(
					awsCheckFunc(appName, secretName2, integrationName, ttl1, roleARN)...,
				),
			},
			// Changing mutable fields causes an update
			{
				Config: awsDynamicSecretConfig(appName, secretName2, integrationName, ttl2, roleARN),
				Check: resource.ComposeTestCheckFunc(
					awsCheckFunc(appName, secretName2, integrationName, ttl2, roleARN)...,
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
				Config: awsDynamicSecretConfig(appName, secretName2, integrationName, ttl2, roleARN),
				Check: resource.ComposeTestCheckFunc(
					awsCheckFunc(appName, secretName2, integrationName, ttl2, roleARN)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if awsDynamicSecretExists(t, secretName1) {
				return fmt.Errorf("test dynamic secret %s was not destroyed", secretName1)
			}
			if awsDynamicSecretExists(t, secretName2) {
				return fmt.Errorf("test dynamic secret %s was not destroyed", secretName2)
			}
			return nil
		},
	})
}

func awsDynamicSecretConfig(appName, name, integrationName, ttl, iamRoleARN string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_dynamic_secret" "acc_test_aws" {
	  app_name = %q
	  secret_provider = "aws"
	  name     = %q
	  integration_name = %q
	  default_ttl = %q
	  aws_assume_role = {
		iam_role_arn = %q
	  }
	}`, appName, name, integrationName, ttl, iamRoleARN)
}

func awsCheckFunc(appName, name, integrationName, ttl, iamRoleARN string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_dynamic_secret.acc_test_aws", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "app_name", appName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "secret_provider", "aws"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "name", name),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "integration_name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "default_ttl", ttl),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_aws", "aws_assume_role.iam_role_arn", iamRoleARN),
	}
}

func awsDynamicSecretExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetAwsDynamicSecret(
		secret_service.NewGetAwsDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Secret != nil
}

func testAccVaultSecretsResourceDynamicSecretGCP(t *testing.T) {
	serviceAccountEmail := checkRequiredEnvVarOrFail(t, "HVS_DYNAMIC_SECRET_SERVICE_ACCOUNT_EMAIL")
	integrationName := checkRequiredEnvVarOrFail(t, "HVS_DYNAMIC_SECRET_INTEGRATION_NAME")
	appName := checkRequiredEnvVarOrFail(t, "HVS_APP_NAME")
	ttl1 := "901s"
	ttl2 := "902s"
	secretName1 := "acc_tests_gcp_1"
	secretName2 := "acc_tests_gcp_2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial dynamic secret
			{
				Config: gcpDynamicSecretConfig(appName, secretName1, integrationName, ttl1, serviceAccountEmail),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFunc(appName, secretName1, integrationName, ttl1, serviceAccountEmail)...,
				),
			},
			// Changing an immutable field causes a recreation
			{
				Config: gcpDynamicSecretConfig(appName, secretName2, integrationName, ttl1, serviceAccountEmail),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFunc(appName, secretName2, integrationName, ttl1, serviceAccountEmail)...,
				),
			},
			// Changing mutable fields causes an update
			{
				Config: gcpDynamicSecretConfig(appName, secretName2, integrationName, ttl2, serviceAccountEmail),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFunc(appName, secretName2, integrationName, ttl2, serviceAccountEmail)...,
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
				Config: gcpDynamicSecretConfig(appName, secretName2, integrationName, ttl2, serviceAccountEmail),
				Check: resource.ComposeTestCheckFunc(
					gcpCheckFunc(appName, secretName2, integrationName, ttl2, serviceAccountEmail)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if gcpDynamicSecretExists(t, secretName1) {
				return fmt.Errorf("test dynamic secret %s was not destroyed", secretName1)
			}
			if gcpDynamicSecretExists(t, secretName2) {
				return fmt.Errorf("test dynamic secret %s was not destroyed", secretName2)
			}
			return nil
		},
	})
}

func gcpDynamicSecretConfig(appName, name, integrationName, ttl, serviceAccountEmail string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_dynamic_secret" "acc_test_gcp" {
	  app_name = %q
	  secret_provider = "gcp"
	  name     = %q
	  integration_name = %q
	  default_ttl = %q
	  gcp_impersonate_service_account = {
		service_account_email = %q
	  }
	}`, appName, name, integrationName, ttl, serviceAccountEmail)
}

func gcpCheckFunc(appName, name, integrationName, ttl, serviceAccountEmail string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "app_name", appName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "secret_provider", "gcp"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "name", name),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "integration_name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "default_ttl", ttl),
		resource.TestCheckResourceAttr("hcp_vault_secrets_dynamic_secret.acc_test_gcp", "gcp_impersonate_service_account.service_account_email", serviceAccountEmail),
	}
}

func gcpDynamicSecretExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecretsPreview.GetGcpDynamicSecret(
		secret_service.NewGetGcpDynamicSecretParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseCodeNotFound(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.Secret != nil
}
