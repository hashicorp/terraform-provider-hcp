---
subcategory: ""
page_title: "Peer an HVN to an HCP Consul cluster - HCP Provider"
description: |-
    An example of peering an HVN to an HCP Consul cluster.
---

# Peer an HVN to an HCP Consul cluster

TODO: resource names, how peering and clusters relate

```terraform
# TODO add comment
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

# TODO add comment
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

# TODO add comment
resource "hcp_aws_network_peering" "example_peering" {
  hvn_id = hcp_hvn.example_hvn.hvn_id

  peer_vpc_id         = aws_vpc.main.id
  peer_account_id     = aws_vpc.main.owner_id
  peer_vpc_region     = data.aws_arn.main.region
  peer_vpc_cidr_block = aws_vpc.main.cidr_block
}
```