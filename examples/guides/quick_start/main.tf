# TODO defaults only, add comment
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

# TODO defaults only, add comment
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

# TODO defaults only, add comment
resource "hcp_aws_network_peering" "example_peering" {
  hvn_id                = hcp_hvn.example_hvn.hvn_id

  target_vpc_id         = aws_vpc.main.id
  target_account_id     = aws_vpc.main.owner_id
  target_vpc_region     = data.aws_arn.main.region
  target_vpc_cidr_block = aws_vpc.main.cidr_block
}