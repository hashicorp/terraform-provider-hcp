package providersdkv2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_Platform_DNSForwardingDataSource_AWS tests the DNS forwarding data source.
func TestAcc_Platform_DNSForwardingRuleDataSource_AWS(t *testing.T) {
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
				Config: testAccDNSForwardingRuleDataSourceConfigAWS(uniqueName),
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

func testAccDNSForwardingRuleDataSourceConfigAWS(uniqueName string) string {
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

// TestAcc_Platform_DNSForwardingRuleDataSource_Azure tests the DNS forwarding rule data source with Azure hvn-peering.
func TestAcc_Platform_DNSForwardingRuleDataSource_Azure(t *testing.T) {
	uniqueName := testAccUniqueNameWithPrefix("dns-r-azure")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {
				Source:            "hashicorp/azurerm",
				VersionConstraint: "~> 3.0",
			},
			"azuread": {
				Source:            "hashicorp/azuread",
				VersionConstraint: "~> 2.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDNSForwardingRuleDataSourceConfigAzure(uniqueName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "state"),
					resource.TestCheckResourceAttrSet("data.hcp_dns_forwarding_rule.test", "self_link"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "rule_id", fmt.Sprintf("%s-initial", uniqueName)),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "domain_name", "azure.internal"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.#", "2"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.0", "10.0.1.10"),
					resource.TestCheckResourceAttr("data.hcp_dns_forwarding_rule.test", "inbound_endpoint_ips.1", "10.0.1.11"),
				),
			},
		},
	})
}

func testAccDNSForwardingRuleDataSourceConfigAzure(uniqueName string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

provider "azuread" {}

resource "hcp_hvn" "test" {
  hvn_id         = "%[1]s"
  cloud_provider = "azure"
  region         = "eastus"
  cidr_block     = "172.25.16.0/20"
}

resource "azurerm_resource_group" "test" {
  name     = "%[1]s"
  location = "East US"
}

resource "azurerm_virtual_network" "test" {
  name                = "%[1]s"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "hcp_azure_peering_connection" "test" {
  hvn_link                 = hcp_hvn.test.self_link
  peering_id               = "%[1]s"
  peer_subscription_id     = "%[2]s"
  peer_tenant_id           = "%[3]s"
  peer_vnet_name           = azurerm_virtual_network.test.name
  peer_resource_group_name = azurerm_resource_group.test.name
  peer_vnet_region         = "eastus"
}

resource "azuread_service_principal" "test" {
  application_id = hcp_azure_peering_connection.test.application_id
}

resource "azurerm_role_definition" "test" {
  name  = "%[1]s"
  scope = azurerm_virtual_network.test.id

  assignable_scopes = [
    azurerm_virtual_network.test.id
  ]

  permissions {
    actions = [
      "Microsoft.Network/virtualNetworks/peer/action",
      "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/read",
      "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write"
    ]
  }
}

resource "azurerm_role_assignment" "test" {
  principal_id       = azuread_service_principal.test.id
  scope              = azurerm_virtual_network.test.id
  role_definition_id = azurerm_role_definition.test.role_definition_resource_id
}

data "hcp_azure_peering_connection" "test" {
  hvn_link              = hcp_hvn.test.self_link
  peering_id            = hcp_azure_peering_connection.test.peering_id
  wait_for_active_state = true

  depends_on = [azurerm_role_assignment.test]
}

resource "hcp_dns_forwarding" "test" {
  hvn_id            = hcp_hvn.test.hvn_id
  dns_forwarding_id = "%[1]s"
  peering_id        = "%[1]s"
  connection_type   = "hvn-peering"

  # Ensure peering is in active state before creating DNS forwarding
  depends_on = [data.hcp_azure_peering_connection.test]

  forwarding_rule {
    rule_id              = "%[1]s-initial"
    domain_name          = "azure.internal"
    inbound_endpoint_ips = ["10.0.1.10", "10.0.1.11"]
  }
}

data "hcp_dns_forwarding_rule" "test" {
  hvn_id                  = hcp_hvn.test.hvn_id
  dns_forwarding_id       = hcp_dns_forwarding.test.dns_forwarding_id
  dns_forwarding_rule_id  = "%[1]s-initial"
}
`, uniqueName, subscriptionID, tenantID)
}
