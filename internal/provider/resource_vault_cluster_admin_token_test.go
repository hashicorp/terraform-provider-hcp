package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

resource "hcp_vault_cluster_admin_token" "test" {
	cluster_id        = hcp_vault_cluster.test.cluster_id
}
`)
)

func TestAccVaultClusterAdminToken(t *testing.T) {
	resourceName := "hcp_vault_cluster_admin_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccVaultClusterAdminTokenConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}
