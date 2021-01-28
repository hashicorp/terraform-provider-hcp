resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

resource "hcp_consul_cluster" "example" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = var.cluster_id
  tier       = "development"
}

resource "hcp_consul_snapshot" "example" {
  cluster_id    = hcp_consul_cluster.example.cluster_id
  snapshot_name = var.snapshot_name
}