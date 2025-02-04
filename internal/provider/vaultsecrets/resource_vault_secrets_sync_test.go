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
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_sync.acc_test_gitlab", "organization_id"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_gitlab", "project_id", os.Getenv("HCP_PROJECT_ID")),
					resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_gitlab", "name", syncName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_sync.acc_test_gitlab", "integration_name", integrationName)),
			},
		},
	})
}

func syncConfig(syncName, integrationName string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_sync" "acc_test_gitlab" {
	  name = %q
	  integration_name = %q
	}`, syncName, integrationName)
}
