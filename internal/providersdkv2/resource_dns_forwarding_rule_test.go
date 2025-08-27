package providersdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardingRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSForwardingRuleResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "domain_name", "example.com"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.#", "2"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.0", "10.0.1.10"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.1", "10.0.1.11"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding_rule.test", "id"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding_rule.test", "state"),
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding_rule.test", "self_link"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hcp_dns_forwarding_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDNSForwardingRuleResourceConfigUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "domain_name", "updated.example.com"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.#", "1"),
					resource.TestCheckResourceAttr("hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.0", "10.0.1.12"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDNSForwardingRuleResourceConfig() string {
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

resource "hcp_dns_forwarding_rule" "test" {
  hvn_id               = hcp_hvn.test.hvn_id
  dns_forwarding_id    = hcp_dns_forwarding.test.id
  domain_name          = "example.com"
  inbound_endpoint_ips = ["10.0.1.10", "10.0.1.11"]
}
`
}

func testAccDNSForwardingRuleResourceConfigUpdate() string {
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

resource "hcp_dns_forwarding_rule" "test" {
  hvn_id               = hcp_hvn.test.hvn_id
  dns_forwarding_id    = hcp_dns_forwarding.test.id
  domain_name          = "updated.example.com"
  inbound_endpoint_ips = ["10.0.1.12"]
}
`
}
