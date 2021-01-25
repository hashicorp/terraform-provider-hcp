provider "hcp" {
  client_id     = var.client_id
  client_secret = var.client_secret
  project_id    = var.project_id
}

resource "hcp_hvn" "example" {
  hvn_id         = "main-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}
