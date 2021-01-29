---
subcategory: ""
page_title: "Peer an AWS VPC to a HashiCorp Virtual Network (HVN) - HCP Provider"
description: |-
    An example of peering an AWS VPC to a HashiCorp Virtual Network (HVN).
---

# Peer an AWS VPC to a HashiCorp Virtual Network (HVN)

In order to connect AWS workloads to an HCP Consul cluster, you must peer the VPC in which the workloads reside to the HVN in which the HCP cluster resides.
This is accomplished by using the `hcp_aws_network_peering` resource to create a Network peering between the HVN's VPC and your own VPC.
The [aws_vpc_peering_connection_accepter](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_peering_connection_accepter) resource is useful for accepting the Network peering that is initiated from the `hcp_aws_network_peering`.
-> **Note** The CIDR blocks of the HVN and the peer VPC cannot overlap.

For more details about deploying Consul on HCP, check out the [Deploy HashiCorp Cloud Platform (HCP) Consul](https://learn.hashicorp.com/tutorials/cloud/consul-deploy?in=consul/cloud) guide.

```terraform
// Create a HashiCorp Virtual Network (HVN).
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
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

// Create an HCP Network peering to peer your HVN with your AWS VPC.
resource "hcp_aws_network_peering" "example" {
  peering_id          = var.peer_id
  hvn_id              = hcp_hvn.example.hvn_id
  peer_vpc_id         = aws_vpc.peer.id
  peer_account_id     = aws_vpc.peer.owner_id
  peer_vpc_region     = var.peer_vpc_region
  peer_vpc_cidr_block = aws_vpc.peer.cidr_block
}

// Accept the VPC peering within your AWS account.
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = hcp_aws_network_peering.example.provider_peering_id
  auto_accept               = true
}
```
