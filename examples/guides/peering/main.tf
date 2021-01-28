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

// Create an HCP Network Peering to peer your HVN with your AWS VPC.
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
