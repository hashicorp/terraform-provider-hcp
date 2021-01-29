resource "hcp_hvn" "example_hvn" {
  hvn_id         = "hcp-tf-example-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_consul_cluster" "example_consul_cluster" {
  hvn_id     = hcp_hvn.example_hvn.hvn_id
  cluster_id = "hcp-tf-example-consul-cluster"
  tier       = "development"
}

resource "hcp_aws_network_peering" "example_peering" {
  hvn_id = hcp_hvn.example_hvn.hvn_id
  peer_vpc_id         = "vpc-2f09a348"
  peer_account_id     = "1234567890"
  peer_vpc_region     = "us-west-2"
  peer_vpc_cidr_block = "10.0.1.0/24"
}