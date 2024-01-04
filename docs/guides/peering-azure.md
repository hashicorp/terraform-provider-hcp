---
subcategory: ""
page_title: "Peer an Azure VNet to a HashiCorp Virtual Network (HVN)"
description: |-
    An example of peering an Azure VNet to a HashiCorp Virtual Network (HVN).
---

# Peer an Azure VNet to a HashiCorp Virtual Network (HVN)

In order to connect Azure workloads to an HCP services, you must peer the VNet in which the workloads reside to the HVN in which the HCP service resides.
This is accomplished by using the `hcp_azure_peering_connection` resource to create a network peering between the HVN's VNet and your own VNet.

-> **Note:** The CIDR blocks of the HVN and the peer VNet cannot overlap.

-> **Note:** Azure Hub/Spoke architecture support is in private beta. Please contact [HashiCorp support](https://support.hashicorp.com/hc/en-us) for details.

```terraform
provider "azurerm" {
  features {}
}

provider "azuread" {}
provider "hcp" {}

resource "hcp_hvn" "hvn" {
  hvn_id         = "main-hvn"
  cloud_provider = "azure"
  region         = "westus2"
  cidr_block     = "172.25.16.0/20"
}

// This resource initially returns in a Pending state, because its application_id is required to complete acceptance of the connection.
resource "hcp_azure_peering_connection" "peer" {
  hvn_link                 = hcp_hvn.hvn.self_link
  peering_id               = "dev"
  peer_vnet_name           = azurerm_virtual_network.vnet.name
  peer_subscription_id     = azurerm_subscription.sub.subscription_id
  peer_tenant_id           = "<tenant UUID>"
  peer_resource_group_name = azurerm_resource_group.rg.name
  peer_vnet_region         = azurerm_virtual_network.vnet.location
}

// This data source is the same as the resource above, but waits for the connection to be Active before returning.
data "hcp_azure_peering_connection" "peer" {
  hvn_link              = hcp_hvn.hvn.self_link
  peering_id            = hcp_azure_peering_connection.peer.peering_id
  wait_for_active_state = true
}

// The route depends on the data source, rather than the resource, to ensure the peering is in an Active state.
resource "hcp_hvn_route" "route" {
  hvn_link         = hcp_hvn.hvn.self_link
  hvn_route_id     = "azure-route"
  destination_cidr = "172.31.0.0/16"
  target_link      = data.hcp_azure_peering_connection.peer.self_link
}

data "azurerm_subscription" "sub" {
  subscription_id = "<subscription UUID>"
}

resource "azurerm_resource_group" "rg" {
  name     = "resource-group-test"
  location = "West US"
}

resource "azurerm_virtual_network" "vnet" {
  name                = "vnet-test"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  address_space = [
    "10.0.0.0/16"
  ]
}

resource "azuread_service_principal" "principal" {
  application_id = hcp_azure_peering_connection.peer.application_id
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
```

## Peer an Azure VNet to an HVN - Gateway support

The following example shows how to connect Azure workloads to HCP HVNs which require [Hub-spoke network topology](https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/hybrid-networking/hub-spoke?tabs=cli) utilizing an Azure VPN Gateway.

Notable aspects of this configuration:
* When the `use_remote_gatways` parameter of the `hcp_azure_peering_connection` resource is set to `true`, the peering link from customer VNet to HVN is set with `AllowGatewayTransit` to `true`.

```terraform
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
```

## Peer an Azure VNet to an HVN - Network Virtual Appliance (NVA) support

The following example shows how to connect Azure workloads to HCP HVNs which require [Hub-spoke network topology](https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/hybrid-networking/hub-spoke?tabs=cli) utilizing an Azure Network Virtual Appliance.

Notable aspects of this configuration:
* An Azure Firewall is used as the Network Virtual Appliance (NVA).
* A "Spoke" Virtual Network is included and peered to the "Hub" Virtual Network.

```terraform
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
  cidr            = "172.25.16.0/20"
}

resource "random_uuid" "uuid" {}

resource "hcp_hvn" "hvn" {
  hvn_id         = "hvn-${local.short_uid}"
  cloud_provider = "azure"
  region         = "eastus"
  cidr_block     = local.cidr
}

resource "hcp_azure_peering_connection" "peering" {
  hvn_link                 = hcp_hvn.hvn.self_link
  peering_id               = "pcx-${local.short_uid}"
  peer_subscription_id     = local.subscription_id
  peer_tenant_id           = local.tenant_id
  peer_vnet_name           = azurerm_virtual_network.vnet.name
  peer_resource_group_name = azurerm_resource_group.rg.name
  peer_vnet_region         = "eastus"

  allow_forwarded_traffic = true
  use_remote_gateways     = false
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
  destination_cidr = azurerm_virtual_network.spoke.address_space[0]
  target_link      = data.hcp_azure_peering_connection.peering.self_link

  azure_config {
    next_hop_type       = "VIRTUAL_APPLIANCE"
    next_hop_ip_address = azurerm_firewall.firewall.ip_configuration[0].private_ip_address
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

resource "azurerm_subnet" "spoke" {
  name                 = "subnet-spoke-${local.short_uid}"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.spoke.name

  address_prefixes = [
    "10.1.0.0/24"
  ]
}

resource "azurerm_virtual_network" "spoke" {
  name                = "vnet-spoke-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  address_space = [
    "10.1.0.0/16"
  ]
}

resource "azurerm_route_table" "spoke" {
  name                = "rtab-spoke-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  disable_bgp_route_propagation = false
}

resource "azurerm_subnet_route_table_association" "spoke" {
  subnet_id      = azurerm_subnet.spoke.id
  route_table_id = azurerm_route_table.spoke.id
}

resource "azurerm_route" "spoke" {
  name                = "route-spoke-${local.short_uid}"
  resource_group_name = azurerm_resource_group.rg.name
  route_table_name    = azurerm_route_table.spoke.name

  address_prefix         = local.cidr
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
  name                = "firewall-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  sku_name = "AZFW_VNet"
  sku_tier = "Standard"

  ip_configuration {
    name                 = "firewall-conf-${local.short_uid}"
    subnet_id            = azurerm_subnet.firewall.id
    public_ip_address_id = azurerm_public_ip.firewall.id
  }
}


resource "azurerm_public_ip" "firewall" {
  name                = "firewall-ip-${local.short_uid}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_firewall_network_rule_collection" "firewall" {
  name                = "firewall-collection-${local.short_uid}"
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
  name                = "peering-sth-${local.short_uid}"
  resource_group_name = azurerm_resource_group.rg.name

  virtual_network_name         = azurerm_virtual_network.spoke.name
  remote_virtual_network_id    = azurerm_virtual_network.vnet.id
  allow_virtual_network_access = true
  allow_forwarded_traffic      = true
  allow_gateway_transit        = false
  use_remote_gateways          = false
}

resource "azurerm_virtual_network_peering" "firewall_hubtospoke" {
  name                = "peering-hts-${local.short_uid}"
  resource_group_name = azurerm_resource_group.rg.name

  virtual_network_name         = azurerm_virtual_network.vnet.name
  remote_virtual_network_id    = azurerm_virtual_network.spoke.id
  allow_virtual_network_access = true
  allow_forwarded_traffic      = false
  allow_gateway_transit        = false
  use_remote_gateways          = false
}

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
```

## Tutorials

Refer to the following tutorials for additional usage examples:

- [Configure Azure VM for HCP](https://developer.hashicorp.com/consul/tutorials/cloud-production/consul-client-azure-virtual-machines)
