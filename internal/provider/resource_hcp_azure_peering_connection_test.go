package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	uniqueId                  = fmt.Sprintf("hcp-tf-provider-test-%d", rand.Intn(99999))
	testAccAzurePeeringConfig = fmt.Sprintf(`
provider "azurem" {
  features {}
}

resource "hcp_hvn" "test" {
  hvn_id         = "hvn-test"
  cloud_provider = "azure"
  region         = "westus2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_azure_peering_connection" "peering" {
  hvn                      = hcp_hvn.test.self_link
  peering_id               = "%[1]s-peering"
  peer_vnet_name           = azurerm_virtual_network.vnet.name
  peer_subscription_id     = "subscription-uuid"
  peer_tenant_id           = "tenant-uuid"
  peer_resource_group_name = azurerm_resource_group.rg.name
  peer_vnet_region         = "westus2"
}

resource "hcp_hvn_route" "route" {
  hvn_route_id = "%[1]s-route"
  hvn_link = hcp_hvn.test.self_link
  destination_cidr = "172.31.0.0/16"
  target_link = hcp_azure_peering_connection.peering.self_link
}

// TODO: peering activation resource

resource "azurerm_resource_group" "rg" {
  name     = "resource-group-test"
  location = "West US"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "%[1]s-vnet"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  address_space = [
    "10.0.0.0/16"
  ]
}

resource "azuread_service_principal" "principal" {
  application_id = hcp_azure_peering_connection.peering.application_id
}

resource "azurerm_role_definition" "definition" {
  name  = "hcp-hvn-peering-access"
  scope = azurerm_virtual_network.vnet.id

  assignable_scopes = [
    azurerm_virtual_network.vnet.id
  ]

  permissions {
    actions = [
      "Microsoft.Network/virtualNetworks/peer/action",
      "Microsoft.Network/virtualNetworks/virtualNetworkPeerings/write"
    ]
  }
}

resource "azurerm_role_assignment" "assignment" {
  principal_id       = azuread_service_principal.principal.id
  scope              = azurerm_virtual_network.vnet.id
  role_definition_id = azurerm_role_definition.definition.role_definition_resource_id
}
`, uniqueId)
)

func TestAccAzurePeeringConnection(t *testing.T) {
	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, true) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azure": {VersionConstraint: "~> 2.46.0"},
		},
		CheckDestroy: testAccCheckHvnPeeringDestroy,

		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccAzurePeeringConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAzurePeeringExists(name string) resource.TestCheckFunc {
	return nil
}

func testAccCheckAzurePeeringDestroy(s *terraform.State) error {
	return nil
}
