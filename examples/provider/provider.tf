// Pin the version
terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.3"
    }
  }
}

// Configure the provider
provider "hcp" {}

// Use the cloud provider AWS to provision resources that will be connected to HCP
provider "aws" {
  region = var.region
}

// Create an HVN
resource "hcp_hvn" "example_hvn" {
  hvn_id         = "hcp-tf-example-hvn"
  cloud_provider = "aws"
  region         = var.region
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

// Create a network peering between the HVN and the AWS VPC
resource "hcp_aws_network_peering" "example_peering" {
  hvn_id              = hcp_hvn.example_hvn.hvn_id
  peer_vpc_id         = aws_vpc.main.id
  peer_account_id     = aws_vpc.main.owner_id
  peer_vpc_region     = data.aws_arn.main.region
  peer_vpc_cidr_block = aws_vpc.main.cidr_block
}

// Create a Consul cluster in the same region and cloud provider as the HVN
resource "hcp_consul_cluster" "example" {
  hvn_id     = hcp_hvn.example_hvn.hvn_id
  cluster_id = "hcp-tf-example-consul-cluster"
  tier       = "development"
}

// Create a secondary Consul cluster to federate with the existing Consul cluster
resource "hcp_consul_cluster" "example_secondary" {
  hvn_id       = hcp_hvn.example_hvn.hvn_id
  cluster_id   = "hcp-tf-example-consul-cluster-secondary"
  tier         = "development"
  primary_link = hcp_consul_cluster.example.self_link
}