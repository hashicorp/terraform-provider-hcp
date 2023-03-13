provider "hcp" {}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}
