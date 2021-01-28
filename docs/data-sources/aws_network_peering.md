---
page_title: "hcp_aws_network_peering Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The AWS Network Peering data source provides information about an existing network peering between an HVN and a peer AWS VPC.
---

# Data Source `hcp_aws_network_peering`

The AWS Network Peering data source provides information about an existing network peering between an HVN and a peer AWS VPC.

## Example Usage

```terraform
data "hcp_aws_network_peering" "test" {
  hvn_id     = var.hvn_id
  peering_id = var.peering_id
}
```

## Schema

### Required

- **hvn_id** (String) The ID of the HashiCorp Virtual Network (HVN).
- **peering_id** (String) The ID of the network peering.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **created_at** (String) The time that the network peering was created.
- **expires_at** (String) The time after which the network peering will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.
- **organization_id** (String) The ID of the HCP organization where the network peering is located. Always matches the HVN's organization.
- **peer_account_id** (String) The account ID of the peer VPC in AWS.
- **peer_vpc_cidr_block** (String) The CIDR range of the peer VPC in AWS.
- **peer_vpc_id** (String) The ID of the peer VPC in AWS.
- **peer_vpc_region** (String) The region of the peer VPC in AWS.
- **project_id** (String) The ID of the HCP project where the network peering is located. Always matches the HVN's project.
- **provider_peering_id** (String) The network peering ID used by AWS.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


