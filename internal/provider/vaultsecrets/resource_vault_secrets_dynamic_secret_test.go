package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceDynamicSecret(t *testing.T) {
	// TODO read env vars for the target provider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{
			// TODO add steps
		},
		CheckDestroy: func(_ *terraform.State) error {
			// TODO check dynamic secrets do not exist anymore
			return nil
		},
	})
}

func awsDynamicSecretConfig(appName, name, integrationName, iamRoleARN string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_dynamic_secret" "acc_test_aws" {
	  app_name = %q
	  secret_provider = "aws"
	  name     = %q
	  integration_name = %q
	  default_ttl = "900s"
	  aws_assume_role = {
		iam_role_arn = %q
	  }
	}`, appName, name, integrationName, iamRoleARN)
}

func gcpDynamicSecretConfig(appName, name, integrationName, serviceAccountEmail string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_dynamic_secret" "acc_test_gcp" {
	  app_name = %q
	  secret_provider = "gcp"
	  name     = %q
	  integration_name = %q
	  default_ttl = "900s"
	  gcp_impersonate_service_account = {
		service_account_email = %q
	  }
	}`, appName, name, integrationName, serviceAccountEmail)
}

func awsCheckFunc(appName, name, integrationName, iamRoleARN string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_dynamic_secret.acc_test_aws", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test_aws", "project_id", os.Getenv("HCP_PROJECT_ID")),
		// TODO check other fields
	}
}

func gcpCheckFunc(appName, name, integrationName, serviceAccountEmail string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_dynamic_secret.acc_test_aws", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_integration_gcp.acc_test_aws", "project_id", os.Getenv("HCP_PROJECT_ID")),
		// TODO check other fields
	}
}
