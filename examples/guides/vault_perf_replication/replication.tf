resource "hcp_hvn" "primary_network" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_vault_cluster" "primary" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.primary_network.hvn_id
  tier       = "plus_medium"
}

resource "hcp_hvn" "secondary_network" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "eu-central-1"
  cidr_block     = "172.26.16.0/20"
}

resource "hcp_vault_cluster" "secondary" {
  cluster_id   = "vault-cluster"
  hvn_id       = hcp_hvn.secondary_network.hvn_id
  tier         = "plus_medium"
  primary_link = hcp_vault_cluster.primary.self_link
}
