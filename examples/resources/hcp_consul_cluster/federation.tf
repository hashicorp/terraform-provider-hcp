data "hcp_hvn" "example" {
  hvn_id = "hvn"
}

data "hcp_consul_cluster" "primary" {
  cluster_id = "consul-cluster-primary"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id       = data.hcp_hvn.example.hvn_id
  cluster_id   = "consul-cluster-secondary"
  tier         = "development"
  primary_link = data.hcp_consul_cluster.primary.self_link
}

