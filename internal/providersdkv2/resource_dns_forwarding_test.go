package providersdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardingResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSForwardingResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding.test", "id"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding.test", "created_at"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding.test", "state"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding.test", "self_link"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding.test", "dns_forwarding_id", "test-dns-forwarding"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding.test", "connection_type", "PEERING"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding.test", "forwarding_rule.#", "1"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding.test", "forwarding_rule.0.domain_name", "example.internal"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hcp_dns_forwarding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDNSForwardingResourceConfig() string {
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
  connection_type   = "PEERING"
  
  forwarding_rule {
    rule_id              = "test-rule"
    domain_name          = "example.internal"
    inbound_endpoint_ips = ["10.0.1.10", "10.0.1.11"]
  }
}
`
}
