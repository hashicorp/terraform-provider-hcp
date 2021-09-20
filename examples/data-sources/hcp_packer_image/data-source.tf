data "hcp_packer_iteration" "hardened-source" {
  bucket_name = "hardened-ubuntu-16-04"
  channel     = "production"
}

data "hcp_packer_image" "foo" {
  bucket_name    = "hardened-ubuntu-16-04"
  cloud_provider = "aws"
  iteration_id   = data.hcp_packer_iteration.hardened-source.ulid
  region         = "us-east-1"
}

output "packer-registry-ubuntu" {
  value = data.hcp_packer_image.foo.cloud_image_id
}
