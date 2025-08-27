package providersdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardingDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSForwardingDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding.test", "id"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding.test", "state"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding.test", "self_link"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding.test", "aws_account_id"),
				),
			},
		},
	})
}

func testAccDNSForwardingDataSourceConfig() string {
	return `
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_dns_forwarding" "test" {
  hvn_id = hcp_hvn.test.hvn_id
}

data "hcp_dns_forwarding" "test" {
  hvn_id = hcp_hvn.test.hvn_id
  depends_on = [hcp_dns_forwarding.test]
}
`
}
