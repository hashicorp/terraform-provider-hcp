---
page_title: "hcp_aws_network_peering Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The AWS Network Peering resource allows you to manage a peering connection between an HVN and a peer AWS VPC.
---

# Resource `hcp_aws_network_peering`

The AWS Network Peering resource allows you to manage a peering connection between an HVN and a peer AWS VPC.

## Example Usage

```terraform
provider "hcp" {
  client_id     = var.client_id
  client_secret = var.client_secret
  project_id    = var.project_id
}

provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "main" {
  hvn_id         = "main-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "aws_vpc" "peer" {
  cidr_block = "172.31.0.0/16"
}

data "aws_arn" "peer" {
  arn = aws_vpc.peer.arn
}

resource "hcp_aws_network_peering" "peer" {
  hvn_id              = hcp_hvn.main.hvn_id
  peer_vpc_id         = aws_vpc.peer.id
  peer_account_id     = aws_vpc.peer.owner_id
  peer_vpc_region     = data.aws_arn.peer.region
  peer_vpc_cidr_block = aws_vpc.peer.cidr_block
}

resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = hcp_aws_network_peering.peer.provider_peering_id
  auto_accept               = true
}
```

## Schema

### Required

- **hvn_id** (String) The ID of the HashiCorp Virtual Network.
- **peer_account_id** (String) The account ID of the peer VPC in AWS.
- **peer_vpc_cidr_block** (String) The CIDR range of the peer VPC in AWS.
- **peer_vpc_id** (String) The ID of the peer VPC in AWS.
- **peer_vpc_region** (String) The region of the peer VPC in AWS.

### Optional

- **id** (String) The ID of this resource.
- **peering_id** (String) The ID of the network peering.
- **project_id** (String) The ID of the HCP project where the network peering is located. Must match the HVN's project.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **created_at** (String) The time that the network peering was created.
- **expires_at** (String) The time after which the network peering will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.
- **organization_id** (String) The ID of the HCP organization where the network peering is located. Always matches the HVN's organization.
- **provider_peering_id** (String) The peering connection ID used by AWS.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)


