// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	// using unique names for resources to make debugging easier
	hvnRouteUniqueName       = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))
	testAccHvnRouteConfigAws = fmt.Sprintf(`
	  provider "aws" {
		region = "us-west-2"
	  }

	  resource "hcp_hvn" "test" {
		  hvn_id         = "%[1]s"
		  cloud_provider = "aws"
		  region         = "us-west-2"
	  }

	  resource "aws_vpc" "vpc" {
		cidr_block = "10.220.0.0/16"
		tags = {
		   Name = "%[1]s"
		}
	  }

	  resource "hcp_aws_network_peering" "peering" {
		peering_id      = "%[1]s"
		hvn_id          = hcp_hvn.test.hvn_id
		peer_account_id = aws_vpc.vpc.owner_id
		peer_vpc_id     = aws_vpc.vpc.id
		peer_vpc_region = "us-west-2"
	  }

	  // This data source is the same as the resource above, but waits for the connection to be Active before returning.
	  data "hcp_aws_network_peering" "peering" {
		hvn_id                = hcp_hvn.test.hvn_id
		peering_id            = hcp_aws_network_peering.peering.peering_id
		wait_for_active_state = true
	  }

	  // The route depends on the data source, rather than the resource, to ensure the peering is in an Active state.
	  resource "hcp_hvn_route" "route" {
		hvn_route_id     = "%[1]s"
		hvn_link         = hcp_hvn.test.self_link
		destination_cidr = "172.31.0.0/16"
		target_link      = data.hcp_aws_network_peering.peering.self_link
	  }

	  resource "aws_vpc_peering_connection_accepter" "peering-accepter" {
		vpc_peering_connection_id = hcp_aws_network_peering.peering.provider_peering_id
		auto_accept               = true
		tags = {
		   Name = "%[1]s"

		   // we need to have these tags here because peering-accepter will turn into
		   // an actual peering which HCP will populate with a set of tags (the ones below).
		   // After succesfull "apply"" test will try to run "plan" operation
		   // to make sure there are no changes to the state and if we don't specify these
		   // tags here then it will fail.
		   hvn_id          = hcp_hvn.test.hvn_id
		   organization_id = hcp_hvn.test.organization_id
		   project_id      = hcp_hvn.test.project_id
		   peering_id      = hcp_aws_network_peering.peering.peering_id
		}
	  }
	  `, hvnRouteUniqueName)
)

// Azure config
func testAccHvnRouteConfigAzure(azConfig, optConfig string) string {
	return fmt.Sprintf(`
	provider "azurerm" {
	  features {}
	}

	resource "hcp_hvn" "hvn" {
	  hvn_id         = "%[1]s"
	  cloud_provider = "azure"
	  region         = "eastus"
	  cidr_block     = "172.25.16.0/20"
	}

	// This resource initially returns in a Pending state, because its application_id is required to complete acceptance of the connection.
	resource "hcp_azure_peering_connection" "peering" {
	  hvn_link                 = hcp_hvn.hvn.self_link
	  peering_id               = "%[1]s"
	  peer_subscription_id     = "%[2]s"
	  peer_tenant_id           = "%[3]s"
	  peer_resource_group_name = azurerm_resource_group.rg.name
	  peer_vnet_name           = azurerm_virtual_network.vnet.name
	  peer_vnet_region         = "eastus"

	  allow_forwarded_traffic = true
	  use_remote_gateways     = true
	}

	// This data source is the same as the resource above, but waits for the connection to be Active before returning.
	data "hcp_azure_peering_connection" "peering" {
	  hvn_link              = hcp_hvn.hvn.self_link
	  peering_id            = hcp_azure_peering_connection.peering.peering_id
	  wait_for_active_state = true
	}

	// The route depends on the data source, rather than the resource, to ensure the peering is in an Active state.
	resource "hcp_hvn_route" "route" {
	  hvn_route_id     = "%[1]s"
	  hvn_link         = hcp_hvn.hvn.self_link
	  target_link      = data.hcp_azure_peering_connection.peering.self_link
	  destination_cidr = "172.31.0.0/16"

	  %[4]s
	}

	resource "azurerm_resource_group" "rg" {
	  name     = "%[1]s"
	  location = "East US"
	}

	resource "azurerm_virtual_network" "vnet" {
	  name                = "vnet-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name

	  address_space = [
		"10.0.0.0/16"
	  ]
	}

	resource "azurerm_subnet" "gateway" {
	  name                 = "GatewaySubnet"
	  resource_group_name  = azurerm_resource_group.rg.name
	  virtual_network_name = azurerm_virtual_network.vnet.name

	  address_prefixes = [
		"10.0.1.0/24"
	  ]
	}

	resource "azurerm_public_ip" "ip" {
	  name                = "ip-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name
	  allocation_method   = "Dynamic"
	}

	resource "azurerm_virtual_network_gateway" "gateway" {
	  name                = "gateway-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name
	  type                = "Vpn"
	  vpn_type            = "RouteBased"
	  sku                 = "Basic"

	  ip_configuration {
		name                          = "ipc-gateway-%[1]s"
		public_ip_address_id          = azurerm_public_ip.ip.id
		private_ip_address_allocation = "Dynamic"
		subnet_id                     = azurerm_subnet.gateway.id
	  }
	}

	%[5]s
	`, hvnRouteUniqueName, subscriptionID, tenantID, azConfig, optConfig)
}

func testAccHvnRouteConfigNVA() string {
	return fmt.Sprintf(`
	resource "azurerm_subnet" "spoke" {
	  name                 = "as-spoke-%[1]s"
	  resource_group_name  = azurerm_resource_group.rg.name
	  virtual_network_name = azurerm_virtual_network.spoke.name

	  address_prefixes = [
		"10.1.0.0/24"
	  ]
	}

	resource "azurerm_virtual_network" "spoke" {
	  name                = "avn-spoke-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name

	  address_space = [
		"10.1.0.0/16"
	  ]
	}

	resource "azurerm_route_table" "spoke" {
	  name                = "art-spoke-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name

	  disable_bgp_route_propagation = false
	}

	resource "azurerm_subnet_route_table_association" "spoke" {
	  subnet_id      = azurerm_subnet.spoke.id
	  route_table_id = azurerm_route_table.spoke.id
	}

	resource "azurerm_route" "spoke" {
	  name                = "ar-spoke-%[1]s"
	  resource_group_name = azurerm_resource_group.rg.name
	  route_table_name    = azurerm_route_table.spoke.name

	  address_prefix         = hcp_hvn.hvn.cidr_block
	  next_hop_type          = "VirtualAppliance"
	  next_hop_in_ip_address = azurerm_firewall.firewall.ip_configuration[0].private_ip_address
	}

	resource "azurerm_subnet" "firewall" {
	  name                 = "AzureFirewallSubnet"
	  resource_group_name  = azurerm_resource_group.rg.name
	  virtual_network_name = azurerm_virtual_network.vnet.name

	  address_prefixes = [
		"10.0.255.0/24"
	  ]
	}

	resource "azurerm_firewall" "firewall" {
	  name                = "af-firewall-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name

	  sku_name = "AZFW_VNet"
	  sku_tier = "Standard"

	  ip_configuration {
		name                 = "aff-ipconf-%[1]s"
		subnet_id            = azurerm_subnet.firewall.id
		public_ip_address_id = azurerm_public_ip.firewall.id
	  }
	}

	resource "azurerm_public_ip" "firewall" {
	  name                = "api-firewall-%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name
	  allocation_method   = "Static"
	  sku                 = "Standard"
	}

	resource "azurerm_firewall_network_rule_collection" "firewall" {
	  name                = "afnrc-firewall-%[1]s"
	  resource_group_name = azurerm_resource_group.rg.name
	  azure_firewall_name = azurerm_firewall.firewall.name
	  priority            = 100
	  action              = "Allow"

	  rule {
		name                  = "HNVtoSpoke"
		protocols             = ["Any"]
		source_addresses      = [hcp_hvn.hvn.cidr_block]
		destination_addresses = azurerm_virtual_network.spoke.address_space
		destination_ports     = ["*"]
	  }

	  rule {
		name                  = "SpokeToHVN"
		protocols             = ["Any"]
		source_addresses      = azurerm_virtual_network.spoke.address_space
		destination_addresses = [hcp_hvn.hvn.cidr_block]
		destination_ports     = ["*"]
	  }
	}

	resource "azurerm_virtual_network_peering" "firewall_spoketohub" {
	  name                = "avnp-sth-%[1]s"
	  resource_group_name = azurerm_resource_group.rg.name

	  virtual_network_name         = azurerm_virtual_network.spoke.name
	  remote_virtual_network_id    = azurerm_virtual_network.vnet.id
	  allow_virtual_network_access = true
	  allow_forwarded_traffic      = true
	  allow_gateway_transit        = false
	  use_remote_gateways          = false
	}

	resource "azurerm_virtual_network_peering" "firewall_hubtospoke" {
	  name                = "avnp-hts-%[1]s"
	  resource_group_name = azurerm_resource_group.rg.name

	  virtual_network_name         = azurerm_virtual_network.vnet.name
	  remote_virtual_network_id    = azurerm_virtual_network.spoke.id
	  allow_virtual_network_access = true
	  allow_forwarded_traffic      = false
	  allow_gateway_transit        = false
	  use_remote_gateways          = false
	}
	`, hvnRouteUniqueName)
}

// hvnRouteAzureAdConfig is the config required to allow HCP to peer from the Remote VNet to HCP HVN
var hvnRouteAzureAdConfig = `
	resource "azuread_service_principal" "principal" {
	  application_id = hcp_azure_peering_connection.peering.application_id
	}

	resource "azurerm_role_definition" "definition" {
	  name  = "hcp-provider-test-role-def"
	  scope = azurerm_virtual_network.vnet.id

	  assignable_scopes = [
		azurerm_virtual_network.vnet.id
	  ]

	  permissions {
		actions = [
		  "Microsoft.Network/virtualNetworks/peer/action",
		  "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/read",
		  "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write"
		]
	  }
	}

	resource "azurerm_role_assignment" "assignment" {
	  principal_id       = azuread_service_principal.principal.id
	  scope              = azurerm_virtual_network.vnet.id
	  role_definition_id = azurerm_role_definition.definition.role_definition_resource_id
	}
`

var azConfigGateway = `
	  azure_config {
	    next_hop_type = "VIRTUAL_NETWORK_GATEWAY"
	  }
`

var azConfigInvalidNextHopType = `
	  azure_config {
	    next_hop_type        = "VIRTUAL_NETWORK_GATEWAY"
		next_hop_ip_address  = "73.35.181.110"
	  }
`

var azConfigNVA = `
	  azure_config {
	    next_hop_type       = "VIRTUAL_APPLIANCE"
	    next_hop_ip_address = azurerm_firewall.firewall.ip_configuration[0].private_ip_address
	  }
`

func TestAccHvnRouteAws(t *testing.T) {
	resourceName := "hcp_hvn_route.route"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": true, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {VersionConstraint: "~> 4.0.0"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccHvnRouteConfigAws),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					resource.TestCheckNoResourceAttr(resourceName, "azure_config.0"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.test"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					routeID := rs.Primary.Attributes["hvn_route_id"]
					return fmt.Sprintf("%s:%s", hvnID, routeID), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccHvnRouteConfigAws),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					resource.TestCheckNoResourceAttr(resourceName, "azure_config.0"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.test"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func TestAccHvnRouteAzure(t *testing.T) {
	resourceName := "hcp_hvn_route.route"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigGateway, hvnRouteAzureAdConfig)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_NETWORK_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.hvn"].Primary.Attributes["hvn_id"]
					routeID := rs.Primary.Attributes["hvn_route_id"]
					return fmt.Sprintf("%s:%s", hvnID, routeID), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigGateway, hvnRouteAzureAdConfig)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_NETWORK_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
		},
	})
}

// Test Azure Route with Gateway architecture
func TestAccHvnRouteAzureGatewayInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")
	resourceName := "hcp_hvn_route.route"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigGateway, "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_NETWORK_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.hvn"].Primary.Attributes["hvn_id"]
					routeID := rs.Primary.Attributes["hvn_route_id"]
					return fmt.Sprintf("%s:%s", hvnID, routeID), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigGateway, "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_NETWORK_GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
		},
	})
}

// Test Azure Config invalid Next Hop Type
func TestAccHvnRouteAzureNextHopTypeValidInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,
		Steps: []resource.TestStep{
			// Testing invalide azure_config based on next_hop_type value
			{
				Config:      testConfig(testAccHvnRouteConfigAzure(azConfigInvalidNextHopType, "")),
				ExpectError: regexp.MustCompile(`azure configuration is invalid: Next hop IP addresses are only allowed in routes where next hop type is VIRTUAL_APPLIANCE`),
			},
		},
	})
}

// Test Azure Route with NVA architecture
func TestAccHvnRouteAzureNVAInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")
	resourceName := "hcp_hvn_route.route"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigNVA, testAccHvnRouteConfigNVA())),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_APPLIANCE"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_config.0.next_hop_ip_address"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.hvn"].Primary.Attributes["hvn_id"]
					routeID := rs.Primary.Attributes["hvn_route_id"]
					return fmt.Sprintf("%s:%s", hvnID, routeID), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccHvnRouteConfigAzure(azConfigNVA, testAccHvnRouteConfigNVA())),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "azure_config.0.next_hop_type", "VIRTUAL_APPLIANCE"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_config.0.next_hop_ip_address"),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", hvnRouteUniqueName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
					testLink(resourceName, "self_link", hvnRouteUniqueName, HVNRouteResourceType, "hcp_hvn.hvn"),
					testLink(resourceName, "target_link", hvnRouteUniqueName, PeeringResourceType, "hcp_hvn.hvn"),
				),
			},
		},
	})
}

func testAccCheckHvnRouteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*clients.Client)

		hvnRouteLink, err := buildLinkFromURL(id, HVNRouteResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build hvnRouteLink for %q: %v", id, err)
		}

		hvnURL, ok := rs.Primary.Attributes["hvn_link"]
		if !ok {
			return fmt.Errorf("hcp_hvn_route doesn't have hvn_link")
		}
		hvnLink, err := parseLinkURL(hvnURL, HvnResourceType)
		if err != nil {
			return fmt.Errorf("failed to parse hvn_link: %w", err)
		}

		hvnRouteID := hvnRouteLink.ID
		loc := hvnRouteLink.Location

		if _, err := clients.GetHVNRoute(context.Background(), client, hvnLink.ID, hvnRouteID, loc); err != nil {
			return fmt.Errorf("unable to get HVN route %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckHvnRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_hvn_route":
			id := rs.Primary.ID

			hvnRouteLink, err := buildLinkFromURL(id, HVNRouteResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build hvnRouteLink for %q: %v", id, err)
			}

			hvnURL, ok := rs.Primary.Attributes["hvn_link"]
			if !ok {
				return fmt.Errorf("hcp_hvn_route doesn't have hvn_link")
			}
			hvnLink, err := parseLinkURL(hvnURL, HvnResourceType)
			if err != nil {
				return fmt.Errorf("failed to parse hvn_link: %w", err)
			}

			hvnRouteID := hvnRouteLink.ID
			loc := hvnRouteLink.Location

			_, err = clients.GetHVNRoute(context.Background(), client, hvnLink.ID, hvnRouteID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
