provider "hcp" {}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  project_id     = "f709ec73-55d4-46d8-897d-816ebba28778"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_consul_cluster" "consul_cluster" {
  cluster_id = "test-cluster"
  hvn_id     = hcp_hvn.test.hvn_id
  project_id = "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
  tier       = "development"
}