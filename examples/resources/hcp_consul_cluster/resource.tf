provider "hcp" {
  client_id     = var.client_id
  client_secret = var.client_secret
  project_id    = var.project_id
}

resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_consul_cluster" "example" {
  cluster_id     = "consul-cluster"
  hvn_id         = hcp_hvn.main.hvn_id
  cloud_provider = hcp_hvn.main.cloud_provider
  region         = hcp_hvn.main.region
}

