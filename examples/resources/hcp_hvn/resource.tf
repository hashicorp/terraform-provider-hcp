resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cidr_block     = "172.25.16.0/20"
  cloud_provider = "aws"
  region         = "us-west-2"
  project_id     = var.project_id
}
