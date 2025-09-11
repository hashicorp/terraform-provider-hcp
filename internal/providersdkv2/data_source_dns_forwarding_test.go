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
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding.test", "dns_forwarding_id", "test-dns-forwarding"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding.test", "connection_type", "hvn-peering"),
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

resource "hcp_aws_network_peering" "test" {
  hvn_id          = hcp_hvn.test.hvn_id
  peering_id      = "test-peering"
  peer_vpc_id     = "vpc-12345678"
  peer_account_id = "123456789012"
  peer_vpc_region = "us-west-2"
}

resource "hcp_dns_forwarding" "test" {
  hvn_id            = hcp_hvn.test.hvn_id
  dns_forwarding_id = "test-dns-forwarding"
  peering_id        = hcp_aws_network_peering.test.peering_id
  connection_type   = "hvn-peering"
  
  forwarding_rule {
    rule_id              = "test-rule"
    domain_name          = "example.internal"
    inbound_endpoint_ips = ["10.0.1.10", "10.0.1.11"]
  }
}

data "hcp_dns_forwarding" "test" {
  hvn_id            = hcp_hvn.test.hvn_id
  dns_forwarding_id = hcp_dns_forwarding.test.dns_forwarding_id
}
`
}
