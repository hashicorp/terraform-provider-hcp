provider "hcp" {
  client_id     = var.client_id
  client_secret = var.client_secret
  project_id    = var.project_id
}

provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "main" {
  hvn_id         = "main_hvn"
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
  hvn_id                = hcp_hvn.main.hvn_id
  target_vpc_id         = aws_vpc.peer.id
  target_account_id     = aws_vpc.peer.owner_id
  target_vpc_region     = data.aws_arn.peer.region
  target_vpc_cidr_block = aws_vpc.peer.cidr_block
}

resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = hcp_aws_network_peering.peer.provider_peering_id
  auto_accept               = true
}
