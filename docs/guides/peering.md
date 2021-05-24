---
subcategory: ""
page_title: "Peer an AWS VPC to a HashiCorp Virtual Network (HVN) - HCP Provider"
description: |-
    An example of peering an AWS VPC to a HashiCorp Virtual Network (HVN).
---

# Peer an AWS VPC to a HashiCorp Virtual Network (HVN)

In order to connect AWS workloads to an HCP Consul cluster, you must peer the VPC in which the workloads reside to the HVN in which the HCP cluster resides.
This is accomplished by using the `hcp_aws_network_peering` resource to create a network peering between the HVN's VPC and your own VPC.
The [aws_vpc_peering_connection_accepter](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_peering_connection_accepter) resource is useful for accepting the network peering that is initiated from the `hcp_aws_network_peering`.

-> **Note:** The CIDR blocks of the HVN and the peer VPC cannot overlap.

For more details about deploying Consul on HCP, check out the [Deploy HashiCorp Cloud Platform (HCP) Consul](https://learn.hashicorp.com/tutorials/cloud/consul-deploy?in=consul/cloud) guide.

```terraform
// Create a HashiCorp Virtual Network (HVN).
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = "aws"
  region         = var.region
  cidr_block     = "172.25.16.0/20"
}

// Create an HCP Consul cluster within the HVN.
resource "hcp_consul_cluster" "example" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = var.cluster_id
  tier       = "development"
}

// If you have not already, create a VPC within your AWS account that will
// contain the workloads you want to connect to your HCP Consul cluster.
// Make sure the CIDR block of the peer VPC does not overlap with the CIDR
// of the HVN.
resource "aws_vpc" "peer" {
  cidr_block = "10.220.0.0/16"
}

// Create an HCP network peering to peer your HVN with your AWS VPC.
resource "hcp_aws_network_peering" "example" {
  peering_id      = var.peer_id
  hvn_id          = hcp_hvn.example.hvn_id
  peer_vpc_id     = aws_vpc.peer.id
  peer_account_id = aws_vpc.peer.owner_id
  peer_vpc_region = var.region
}

// Create an HVN route that targets your HCP network peering and matches your AWS VPC's CIDR block
resource "hcp_hvn_route" "example" {
  hvn_link         = hcp_hvn.hvn.self_link
  hvn_route_id     = var.route_id
  destination_cidr = aws_vpc.peer.cidr_block
  target_link      = hcp_aws_network_peering.example.self_link
}

// Accept the VPC peering within your AWS account.
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = hcp_aws_network_peering.example.provider_peering_id
  auto_accept               = true
}
```
