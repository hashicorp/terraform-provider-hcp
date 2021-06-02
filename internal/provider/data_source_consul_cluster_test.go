package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataSourceConsulClusterConfig = `
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "hcp_consul_cluster" "test" {
	cluster_id = "test-consul-cluster"
	hvn_id     = hcp_hvn.test.hvn_id
	tier       = "development"
}

data "hcp_consul_cluster" "test" {
	cluster_id = hcp_consul_cluster.test.cluster_id
}
`

func TestAccDataSourceConsulCluster(t *testing.T) {
	resourceName := "hcp_consul_cluster.test"
	dataSourceName := "data.hcp_consul_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testDataSourceConsulClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "public_endpoint", dataSourceName, "public_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "datacenter", dataSourceName, "datacenter"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_enabled", dataSourceName, "connect_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_automatic_upgrades", dataSourceName, "consul_automatic_upgrades"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_snapshot_interval", dataSourceName, "consul_snapshot_interval"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_snapshot_retention", dataSourceName, "consul_snapshot_retention"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_config_file", dataSourceName, "consul_config_file"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_ca_file", dataSourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_version", dataSourceName, "consul_version"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_public_endpoint_url", dataSourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_private_endpoint_url", dataSourceName, "consul_private_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "scale", dataSourceName, "scale"),
					resource.TestCheckResourceAttrPair(resourceName, "tier", dataSourceName, "tier"),
					resource.TestCheckResourceAttrPair(resourceName, "size", dataSourceName, "size"),
					resource.TestCheckResourceAttrPair(resourceName, "self_link", dataSourceName, "self_link"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_link", dataSourceName, "primary_link"),
				),
			},
		},
	})
}
