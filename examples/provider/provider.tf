// Configure the provider
provider "hcp" {
  client_id     = "example-id"
  client_secret = "example-secret"
}

// Use your desired cloud provider to provision resources that will be connected to HCP
provider "aws" {
  region = "us-west-2"
}

// Create an HVN
resource "hcp_hvn" "example_hvn" {
  hvn_id         = "hcp-tf-example-hvn"
  cloud_provider = "aws"
  region         = data.aws_arn.main.region
  cidr_block     = "172.25.16.0/20"
}

// Create a VPC for the HVN to peer into
resource "aws_vpc" "main" {
  cidr_block = "172.25.0.0/20"
}

data "aws_arn" "main" {
  arn = aws_vpc.main.arn
}

resource "aws_vpc_peering_connection_accepter" "main" {
  vpc_peering_connection_id = hcp_aws_network_peering.example_peering.provider_peering_id
  auto_accept               = true
}

// Create a network peering connection between the HVN and the VPC
resource "hcp_aws_network_peering" "example_peering" {
  hvn_id                = hcp_hvn.example_hvn.hvn_id

  target_vpc_id         = aws_vpc.main.id
  target_account_id     = aws_vpc.main.owner_id
  target_vpc_region     = data.aws_arn.main.region
  target_vpc_cidr_block = aws_vpc.main.cidr_block
}

// Create a Consul cluster
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}