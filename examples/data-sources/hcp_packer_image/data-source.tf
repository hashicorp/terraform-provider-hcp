data "hcp_packer_iteration" "hardened-source" {
  bucket_name = "hardened-ubuntu-16-04"
  channel     = "production-stable"
}

data "hcp_packer_image" "foo" {
  bucket_name    = "hardened-ubuntu-16-04"
  cloud_provider = "aws"
  iteration_id   = data.hcp_packer_iteration.hardened-source.id
  region         = "us-east-1"
}

output "packer-registry-ubuntu" {
  value = data.hcp_packer_image.foo.id
}
