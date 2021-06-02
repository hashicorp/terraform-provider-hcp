package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testDataSourceHvnConfig = `
resource "hcp_hvn" "example_hvn" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}

data "hcp_hvn" "test_hvn" {
	hvn_id = hcp_hvn.example_hvn.hvn_id
}
`

func TestAccDataSourceHvn(t *testing.T) {
	resourceName := "hcp_hvn.example_hvn"
	dataSourceName := "data.hcp_hvn.test_hvn"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testDataSourceHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", dataSourceName, "cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_account_id", dataSourceName, "provider_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "self_link", dataSourceName, "self_link"),
				),
			},
		},
	})
}
