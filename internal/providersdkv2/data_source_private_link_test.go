// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Platform_DataSourcePrivateLink(t *testing.T) {
	resourceName := "hcp_private_link.test"
	dataSourceName := "data.hcp_private_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPrivateLinkDestroy,
		Steps: []resource.TestStep{
			// Create the resources (HVN, Vault cluster, private link)
			{
				Config: testAccDataSourcePrivateLinkConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateLinkExists(resourceName),
				),
			},
			// Verify data source
			{
				Config: testAccDataSourcePrivateLinkConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "private_link_id", resourceName, "private_link_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hvn_id", resourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vault_cluster_id", resourceName, "vault_cluster_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "external_name", resourceName, "external_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consumer_regions.#", resourceName, "consumer_regions.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consumer_accounts.#", resourceName, "consumer_accounts.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_link", resourceName, "self_link"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_region", resourceName, "default_region"),
				),
			},
		},
	})
}

func testAccDataSourcePrivateLinkConfig() string {
	return `
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
  hvn_id     = hcp_hvn.test.hvn_id
  cluster_id = "test-vault-cluster"
  tier       = "standard_small"
  public_endpoint = false
  
  major_version_upgrade_config {
    upgrade_type = "AUTOMATIC"
  }
}

resource "hcp_private_link" "test" {
  hvn_id           = hcp_hvn.test.hvn_id
  private_link_id  = "test-private-link"
  vault_cluster_id = hcp_vault_cluster.test.cluster_id

  consumer_accounts = [
    "arn:aws:iam::311485635366:root"
  ]
  
  consumer_regions = [
    "us-west-2"
  ]
}

data "hcp_private_link" "test" {
  hvn_id           = hcp_hvn.test.hvn_id
  private_link_id  = hcp_private_link.test.private_link_id
}
`
}
