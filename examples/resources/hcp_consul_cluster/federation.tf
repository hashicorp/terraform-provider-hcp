resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_consul_cluster" "primary" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = "consul-cluster-primary"
  tier       = "development"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id                  = hcp_hvn.example.hvn_id
  cluster_id              = "consul-cluster-secondary"
  tier                    = "development"
  primary_link            = hcp_consul_cluster.primary.self_link
  auto_hvn_to_hvn_peering = true
}