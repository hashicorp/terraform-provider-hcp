package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceSync(t *testing.T) {
	syncName := generateRandomSlug()
	integrationName := generateRandomSlug()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncConfig(syncName, integrationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_integration.acc_test", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_integration.acc_test", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_aws", "name", syncName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_aws", "integration_name", integrationName),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_sync.acc_test_aws", "organization_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_sync.acc_test_aws", "project_id")),
			},
		},
	})
}

func syncConfig(syncName, integrationName string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_sync" "acc_test_aws" {
	  name = %q
	  integration_name = %q
	}`, syncName, integrationName)
}
