provider "hcp" {
  project_id = "f709ec73-55d4-46d8-897d-816ebba28778"
}

# This HVN will be created in the project "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  project_id     = "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
  cloud_provider = "aws"
  region         = "us-west-2"
}
