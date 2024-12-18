package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceSync(t *testing.T) {
	integrationName := checkRequiredEnvVarOrFail(t, "HVS_DYNAMIC_SECRET_INTEGRATION_NAME")
	// appName := checkRequiredEnvVarOrFail(t, "HVS_APP_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncConfig("acc_test_aws", integrationName),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func syncConfig(integrationName, integrationType string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_sync" "acc_test_aws" {
	  integration_name = %q
	  type = %q
	}`, integrationName, integrationType)
}

func checkSync(name, integrationName, integrationType string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_sync.acc_test_aws", "organization_id"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_aws", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_aws", "integration_name", integrationName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_aws", "type", integrationType),
	}
}
