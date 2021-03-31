package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccVaultClusterAdminTokenConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}
	
resource "hcp_vault_cluster" "test" {
	cluster_id            = "test-vault-cluster"
	hvn_id                = hcp_hvn.test.hvn_id
}

data "hcp_vault_cluster_admin_token" "test" {
	cluster_id        = hcp_vault_cluster.test.cluster_id
}
`)
)

func TestAccVaultClusterAdminToken(t *testing.T) {
	dataSourceName := "data.hcp_vault_cluster_admin_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccVaultClusterAdminTokenConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccVaultClusterAdminToken(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttrSet(dataSourceName, "token"),
				),
			},
		},
	})
}

func testAccVaultClusterAdminToken(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}
