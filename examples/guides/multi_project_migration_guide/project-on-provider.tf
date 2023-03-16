provider "hcp" {
  project_id = "f709ec73-55d4-46d8-897d-816ebba28778"
}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}
