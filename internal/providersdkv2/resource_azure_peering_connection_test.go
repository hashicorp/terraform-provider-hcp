// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	uniqueAzurePeeringTestID = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))
	subscriptionID           = os.Getenv("ARM_SUBSCRIPTION_ID")
	tenantID                 = os.Getenv("ARM_TENANT_ID")
)

// peeringHubSpokeNVAConfig is the hcp_azure_peering_connection config params
// to enable a Hub and Spoke archtecture in the NVA model.
var peeringHubSpokeNVAConfig = `
	  allow_forwarded_traffic = true
	  use_remote_gateways     = false
`

// peeringHubSpokeGatewayConfig is the hcp_azure_peering_connection config params
// to enable a Hub and Spoke archtecture using Gateway transit.
var peeringHubSpokeGatewayConfig = `
	  allow_forwarded_traffic = false
	  use_remote_gateways     = true
`

// peeringHubSpokeNVAandGatewayConfig is the hcp_azure_peering_connection config
// params to enable a Hub and Spoke archtecture supporting both NVA and Gateway.
var peeringHubSpokeNVAandGatewayConfig = `
	  allow_forwarded_traffic = true
	  use_remote_gateways     = true
`

// gatewayConfig is the additional components required for Hub and Spoke architecture
// using a Gateway.
func gatewayConfig() string {
	return fmt.Sprintf(`
	resource "azurerm_subnet" "subnet" {
	  name                 = "GatewaySubnet"
	  resource_group_name  = azurerm_resource_group.rg.name
	  virtual_network_name = azurerm_virtual_network.vnet.name
	  address_prefixes     = ["10.0.1.0/24"]
	}

	resource "azurerm_public_ip" "ip" {
	  name                = "%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name
	  allocation_method   = "Dynamic"
	}

	resource "azurerm_virtual_network_gateway" "gateway" {
	  name                = "%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name
	  type                = "Vpn"
	  enable_bgp          = false
      sku                 = "Basic"

	  ip_configuration {
		name                          = "%[1]s"
		public_ip_address_id          = azurerm_public_ip.ip.id
		private_ip_address_allocation = "Dynamic"
		subnet_id                     = azurerm_subnet.subnet.id
	  }
	}
	`, uniqueAzurePeeringTestID)
}

// azureAdConfig is the config required to allow HCP to peer from the Remote VNet to HCP HVN
var azureAdConfig = `
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

// standardConfig is the configuration for a basic Azure peering.
func standardConfig() string {
	return baseConfig("", azureAdConfig)
}

// nvaConfigWithAd is the configuration for a Hub / Spoke architecture using the
// NVA model, including Service Principal components.
func nvaConfigWithAd() string {
	return baseConfig(peeringHubSpokeNVAConfig, azureAdConfig)
}

// gatewayConfigWithAd is the configuration for a Hub / Spoke architecture using
// the Gateway model, including Service Principal components.
func gatewayConfigWithAd() string {
	return baseConfig(peeringHubSpokeGatewayConfig, fmt.Sprintf(`
    %s

    %s
    `, azureAdConfig, gatewayConfig()))
}

// nvaGatewayConfigWithAd is the configuration for a Hub / Spoke architecture using
// NVA and the Gateway model, including Service Principal components.
func nvaGatewayConfigWithAd() string {
	return baseConfig(peeringHubSpokeNVAandGatewayConfig, fmt.Sprintf(`
    %s

    %s
    `, azureAdConfig, gatewayConfig()))
}

// baseConfig is the config excluding the authorization components (SP, Role, Role assignment).
// This is used to support HashiCorp internal engineers.
func baseConfig(hubSpokeConfig, optConfig string) string {
	return fmt.Sprintf(`
	provider "azurerm" {
	  features {}
	}

	resource "hcp_hvn" "test" {
	  hvn_id         = "%[1]s"
	  cloud_provider = "azure"
	  region         = "eastus"
	  cidr_block     = "172.25.16.0/20"
	}

	// This resource initially returns in a Pending state, because its application_id is required to complete acceptance of the connection.
	resource "hcp_azure_peering_connection" "peering" {
	  hvn_link                 = hcp_hvn.test.self_link
	  peering_id               = "%[1]s"
	  peer_vnet_name           = azurerm_virtual_network.vnet.name
	  peer_subscription_id     = "%[2]s"
	  peer_tenant_id           = "%[3]s"
	  peer_resource_group_name = azurerm_resource_group.rg.name
	  peer_vnet_region         = "eastus"

	  // Hub/Spoke networking config
      %[4]s
	}

	// This data source is the same as the resource above, but waits for the connection to be Active before returning.
	data "hcp_azure_peering_connection" "peering" {
	  hvn_link                 = hcp_hvn.test.self_link
	  peering_id               = hcp_azure_peering_connection.peering.peering_id
	  wait_for_active_state    = true
	}

	// The route depends on the data source, rather than the resource, to ensure the peering is in an Active state.
	resource "hcp_hvn_route" "route" {
	  hvn_route_id = "%[1]s"
	  hvn_link = hcp_hvn.test.self_link
	  destination_cidr = "172.31.0.0/16"
	  target_link = data.hcp_azure_peering_connection.peering.self_link
	}

	resource "azurerm_resource_group" "rg" {
	  name     = "%[1]s"
	  location = "East US"
	}

	resource "azurerm_virtual_network" "vnet" {
	  name                = "%[1]s"
	  location            = azurerm_resource_group.rg.location
	  resource_group_name = azurerm_resource_group.rg.name

	  address_space = [
		"10.0.0.0/16"
	  ]
	}

	%[5]s
	`, uniqueAzurePeeringTestID, subscriptionID, tenantID, hubSpokeConfig, optConfig)
}

func TestAccAzurePeeringConnection(t *testing.T) {
	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,

		Steps: []resource.TestStep{
			{
				// Tests create
				Config: testConfig(standardConfig()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
					// Note: azure_peering_id is not set until the peering is accepted after creation.
				),
			},
			{
				// Tests create / Enables Hub/Spoke with NVA connectivity
				Config: testConfig(nvaConfigWithAd()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			{
				// Tests create / Enables Hub/Spoke with Gateway connectivity
				Config: testConfig(gatewayConfigWithAd()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			{
				// Tests create / Enables Hub/Spoke with NVA and Gateway connectivity
				Config: testConfig(nvaGatewayConfigWithAd()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(standardConfig()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

// TestAccAzurePeeringConnectionStandardInternal is almost identical to the standard
// TestAccAzurePeeringConnection test, but does not include the Azure Service Principal
// components which allow permission for HCP to peer from the "Customer" account to HCP dataplane.
// This modified test exists for HashiCorp internal contributors to adhere to on-demand
// service principal creation via doormat.
func TestAccAzurePeeringConnectionStandardInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")

	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,
		Steps: []resource.TestStep{
			{
				// Tests create
				Config: testConfig(baseConfig("", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(baseConfig("", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func TestAccAzurePeeringConnectionNVAInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")

	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,
		Steps: []resource.TestStep{
			{
				// Tests create / Enables Hub/Spoke with NVA connectivity
				Config: testConfig(baseConfig(peeringHubSpokeNVAConfig, "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(baseConfig("", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func TestAccAzurePeeringConnectionGatewayInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")

	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,
		Steps: []resource.TestStep{
			{
				// Tests create - Enables Hub/Spoke with Gateway transit
				Config: testConfig(baseConfig(peeringHubSpokeGatewayConfig, gatewayConfig())),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(baseConfig("", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func TestAccAzurePeeringConnectionNVAGatewayInternal(t *testing.T) {
	t.Skip("This should not be run on CI, only locally.")

	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 3.63"},
			"azuread": {VersionConstraint: "~> 2.39"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,
		Steps: []resource.TestStep{
			{
				// Tests create - Enables Hub/Spoke with NVA and Gateway transit
				Config: testConfig(baseConfig(peeringHubSpokeNVAandGatewayConfig, gatewayConfig())),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(baseConfig("", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttr(resourceName, "allow_forwarded_traffic", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_remote_gateways", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func testAccCheckAzurePeeringExists(name string) resource.TestCheckFunc {
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

		peeringLink, err := buildLinkFromURL(id, PeeringResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build peeringLink for %q: %v", id, err)
		}

		hvnUrn, ok := rs.Primary.Attributes["hvn_link"]
		if !ok {
			return fmt.Errorf("no hvn_link is set")
		}

		hvnLink, err := buildLinkFromURL(hvnUrn, HvnResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to parse hvn_link link URL for %q: %v", id, err)
		}

		azurePeeringID := peeringLink.ID
		loc := peeringLink.Location

		if _, err := clients.GetPeeringByID(context.Background(), client, azurePeeringID, hvnLink.ID, loc); err != nil {
			return fmt.Errorf("unable to get peering connection %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckAzurePeeringDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_azure_peering_connection":
			id := rs.Primary.ID

			if id == "" {
				return fmt.Errorf("no ID is set")
			}

			peeringLink, err := buildLinkFromURL(id, PeeringResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build peeringLink for %q: %v", id, err)
			}

			hvnUrn, ok := rs.Primary.Attributes["hvn_link"]
			if !ok {
				return fmt.Errorf("no hvn_link is set")
			}

			hvnLink, err := buildLinkFromURL(hvnUrn, HvnResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to parse hvn_link link URL for %q: %v", id, err)
			}

			azurePeeringID := peeringLink.ID
			loc := peeringLink.Location

			_, err = clients.GetPeeringByID(context.Background(), client, azurePeeringID, hvnLink.ID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}

	return nil
}
