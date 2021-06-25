resource "hcp_hvn" "primary" {
  hvn_id         = var.primary_hvn_id
  cloud_provider = var.cloud_provider
  region         = var.primary_region
}

resource "hcp_hvn" "secondary" {
  hvn_id         = var.secondary_hvn_id
  cloud_provider = var.cloud_provider
  region         = var.secondary_region
}

resource "hcp_consul_cluster" "primary" {
  hvn_id     = hcp_hvn.primary.hvn_id
  cluster_id = var.primary_cluster_id
  tier       = "development"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id       = hcp_hvn.secondary.hvn_id
  cluster_id   = var.secondary_cluster_id
  tier         = "development"
  primary_link = hcp_consul_cluster.primary.self_link
}