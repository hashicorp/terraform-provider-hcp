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
					resource.TestCheckResourceAttrSet("hcp_dns_forwarding.test", "aws_account_id"),
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

resource "hcp_dns_forwarding" "test" {
  hvn_id = hcp_hvn.test.hvn_id
}
`
}
