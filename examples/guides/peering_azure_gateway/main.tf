terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.78.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=3.0.0"
    }
  }
}

provider "azurerm" {
  features {}
}

variable "subscription_id" {
  type        = string
  description = "The Azure AD Subscription ID"
}

variable "tenant_id" {
  type        = string
  description = "The Azure AD Tenant ID"
}

locals {
  uid             = random_uuid.uuid.result
  short_uid       = format("%.10s", local.uid)
  subscription_id = var.subscription_id
  tenant_id       = var.tenant_id
}

resource "random_uuid" "uuid" {}

resource "hcp_hvn" "hvn" {
  hvn_id         = "hvn-${local.short_uid}"
  cloud_provider = "azure"
  region         = "eastus"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_azure_peering_connection" "peering" {
  hvn_link                 = hcp_hvn.hvn.self_link
  peering_id               = "pcx-${local.short_uid}"
  peer_subscription_id     = local.subscription_id
  peer_tenant_id           = local.tenant_id
  peer_vnet_name           = azurerm_virtual_network.vnet.name
  peer_resource_group_name = azurerm_resource_group.rg.name
  peer_vnet_region         = "eastus"

  allow_forwarded_traffic = false
  use_remote_gateways     = true
}

// This data source is the same as the resource above, but waits for the connection
// to be Active before returning.
data "hcp_azure_peering_connection" "peering" {
  hvn_link              = hcp_hvn.hvn.self_link
  peering_id            = hcp_azure_peering_connection.peering.peering_id
  wait_for_active_state = true
}

// The route depends on the data source, rather than the resource, to ensure the
// peering is in an Active state.
resource "hcp_hvn_route" "route" {
  hvn_route_id     = "route-${local.short_uid}"
  hvn_link         = hcp_hvn.hvn.self_link
  destination_cidr = "172.31.0.0/16"
  target_link      = data.hcp_azure_peering_connection.peering.self_link

  azure_config {
    next_hop_type = "VIRTUAL_NETWORK_GATEWAY"
  }
}

resource "azurerm_resource_group" "rg" {
  name     = "rg-${local.short_uid}"
  location = "East US"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "vnet-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  address_space = [
    "10.0.0.0/16"
  ]
}

resource "azurerm_subnet" "subnet" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name

  address_prefixes = [
    "10.0.1.0/24"
  ]
}

resource "azurerm_public_ip" "ip" {
  name                = "ip-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Dynamic"
}

resource "azurerm_virtual_network_gateway" "gateway" {
  name                = "gateway-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  type                = "Vpn"
  vpn_type            = "RouteBased"
  enable_bgp          = false // Explicit; defaults to false
  sku                 = "Basic"

  ip_configuration {
    name                          = "ipconf-${local.short_uid}"
    public_ip_address_id          = azurerm_public_ip.ip.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurerm_subnet.subnet.id
  }
}

// The principal deploying the ``azuread_service_principal`` resource below requires
// API Permissions as described here: https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal.
// The principal deploying the ``azurerm_role_definition`` and ``azurerm_role_assigment``
// resources must have Owner or User Access Administrator permissions over an appropriate
// scope that includes your Virtual Network.
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
