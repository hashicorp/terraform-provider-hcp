resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

resource "hcp_consul_cluster" "primary" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = var.primary_cluster_id
  tier       = "development"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id       = hcp_hvn.example.hvn_id
  cluster_id   = var.secondary_cluster_id
  tier         = "development"
  primary_link = hcp_consul_cluster.primary.self_link
}