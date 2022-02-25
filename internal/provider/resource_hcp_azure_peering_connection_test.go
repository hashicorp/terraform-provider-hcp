package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	uniqueAzurePeeringTestID  = fmt.Sprintf("hcp-tf-provider-test-%d", rand.Intn(99999))
	subscriptionID            = os.Getenv("AZURE_SUBSCRIPTION_ID")
	tenantID                  = os.Getenv("AZURE_TENANT_ID")
	testAccAzurePeeringConfig = fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "hcp_hvn" "test" {
  hvn_id         = "%[1]s"
  cloud_provider = "azure"
  region         = "eastus"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_azure_peering_connection" "peering" {
  hvn_link                 = hcp_hvn.test.self_link
  peering_id               = "%[1]s"
  peer_vnet_name           = azurerm_virtual_network.vnet.name
  peer_subscription_id     = "subscription-uuid"
  peer_tenant_id           = "tenant-uuid"
  peer_resource_group_name = azurerm_resource_group.rg.name
  peer_vnet_region         = "eastus"
}

data "hcp_peering_activation" "activation" {
	peering_id = hcp_azure_peering_connection.peering.peering_id
	hvn_link   = hcp_hvn.test.self_link
}

resource "hcp_hvn_route" "route" {
  hvn_route_id = "%[1]s"
  hvn_link = hcp_hvn.test.self_link
  destination_cidr = "172.31.0.0/16"
  target_link = hcp_azure_peering_connection.peering.self_link

  depends_on = [ data.hcp_peering_activation.activation ]
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

resource "azuread_service_principal" "principal" {
  application_id = hcp_azure_peering_connection.peering.application_id
}

resource "azurerm_role_definition" "definition" {
  name  = "%[1]s"
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
`, uniqueAzurePeeringTestID, subscriptionID, tenantID)
)

func TestAccAzurePeeringConnection(t *testing.T) {
	resourceName := "hcp_azure_peering_connection.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": true}) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {VersionConstraint: "~> 2.46.0"},
			"azuread": {VersionConstraint: "~> 2.18.0"},
		},
		CheckDestroy: testAccCheckAzurePeeringDestroy,

		Steps: []resource.TestStep{
			{
				// Tests create
				Config: testConfig(testAccAzurePeeringConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					testLink(resourceName, "self_link", uniqueAzurePeeringTestID, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Tests import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     uniqueAzurePeeringTestID + ":" + uniqueAzurePeeringTestID,
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(testAccAzurePeeringConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzurePeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", uniqueAzurePeeringTestID),
					testLink(resourceName, "hvn_link", uniqueAzurePeeringTestID, HvnResourceType, resourceName),
					resource.TestCheckResourceAttr(resourceName, "peer_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr(resourceName, "peer_tenant_id", tenantID),
					resource.TestCheckResourceAttr(resourceName, "peer_vnet_name", uniqueAzurePeeringTestID),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vnet_region"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
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
