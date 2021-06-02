package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataSourceVaultClusterConfig = `
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
	cluster_id = "test-vault-cluster"
	hvn_id     = hcp_hvn.test.hvn_id
}

data "hcp_vault_cluster" "test" {
	cluster_id = hcp_vault_cluster.test.cluster_id
}
`

func TestAccDataSourceVaultCluster(t *testing.T) {
	resourceName := "hcp_vault_cluster.test"
	dataSourceName := "data.hcp_vault_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testDataSourceVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "public_endpoint", dataSourceName, "public_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "min_vault_version", dataSourceName, "min_vault_version"),
					resource.TestCheckResourceAttrPair(resourceName, "tier", dataSourceName, "tier"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace", dataSourceName, "namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_version", dataSourceName, "vault_version"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_public_endpoint_url", dataSourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_private_endpoint_url", dataSourceName, "vault_private_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
				),
			},
		},
	})
}
