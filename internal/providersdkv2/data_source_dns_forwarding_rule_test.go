package providersdkv2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardingRuleDataSource(t *testing.T) {
	uniqueName := testAccUniqueNameWithPrefix("dns-r")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": true, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {
				Source:            "hashicorp/aws",
				VersionConstraint: "~> 5.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDNSForwardingRuleDataSourceConfig(uniqueName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "state"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "self_link"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "rule_id", fmt.Sprintf("%s-initial", uniqueName)),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "domain_name", "initial.internal"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.#", "2"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.0", "10.220.1.1"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.1", "10.220.1.2"),
				),
			},
		},
	})
}

func testAccDNSForwardingRuleDataSourceConfig(uniqueName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "hcp_hvn" "test" {
  hvn_id         = "%[1]s"
  cloud_provider = "aws"
  region         = "us-east-1"
  cidr_block     = "172.25.16.0/20"
}

resource "aws_ec2_transit_gateway" "test" {
  description = "%[1]s"
  tags = {
    Name = "%[1]s"
  }
}

# Create a resource share for the transit gateway
resource "aws_ram_resource_share" "test" {
  name                      = "%[1]s"
  allow_external_principals = true
  
  tags = {
    Name = "%[1]s"
  }
}

# Associate the HCP account as a principal
resource "aws_ram_principal_association" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  principal          = hcp_hvn.test.provider_account_id
}

# Associate the transit gateway with the resource share
resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "hcp_aws_transit_gateway_attachment" "test" {
  depends_on = [
    aws_ram_principal_association.test,
    aws_ram_resource_association.test,
  ]
  
  hvn_id                        = hcp_hvn.test.hvn_id
  transit_gateway_attachment_id = "%[1]s"
  transit_gateway_id            = aws_ec2_transit_gateway.test.id
  resource_share_arn            = aws_ram_resource_share.test.arn
}

# Accept the Transit Gateway attachment
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = hcp_aws_transit_gateway_attachment.test.provider_transit_gateway_attachment_id
  tags = {
    Name = "%[1]s"
  }
}

# This data source waits for the transit gateway attachment to be Active before returning
data "hcp_aws_transit_gateway_attachment" "test" {
  hvn_id                        = hcp_hvn.test.hvn_id
  transit_gateway_attachment_id = "%[1]s"
  wait_for_active_state         = true
  
  # Ensure the AWS accepter runs before checking for active state
  depends_on = [aws_ec2_transit_gateway_vpc_attachment_accepter.test]
}

resource "hcp_dns_forwarding" "test" {
  hvn_id            = hcp_hvn.test.hvn_id
  dns_forwarding_id = "%[1]s"
  peering_id        = "%[1]s"
  connection_type   = "tgw-attachment"
  
  # Ensure transit gateway attachment is active before creating DNS forwarding
  depends_on = [data.hcp_aws_transit_gateway_attachment.test]
  
  forwarding_rule {
    rule_id              = "%[1]s-initial"
    domain_name          = "initial.internal"
    inbound_endpoint_ips = ["10.220.1.1", "10.220.1.2"]
  }
}

data "hcp_dns_forwarding_rule" "test" {
  hvn_id                  = hcp_hvn.test.hvn_id
  dns_forwarding_id       = hcp_dns_forwarding.test.dns_forwarding_id
  dns_forwarding_rule_id  = "%[1]s-initial"
}
`, uniqueName)
}
