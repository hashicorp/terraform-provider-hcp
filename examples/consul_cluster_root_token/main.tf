resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

// The root_token_accessor_id and root_token_secret_id properties will
// no longer be valid after the new root token is created below
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

// Create a new ACL root token
resource "hcp_consul_cluster_root_token" "example" {
  cluster_id = hcp_consul_cluster.example.cluster_id
}