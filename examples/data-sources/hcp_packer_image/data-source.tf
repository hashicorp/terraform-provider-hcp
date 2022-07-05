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

data "hcp_packer_image" "bar" {
  bucket_name    = "hardened-ubuntu-16-04"
  cloud_provider = "aws"
  iteration_id   = data.hcp_packer_iteration.hardened-source.ulid
  region         = "us-west-1"
}

output "packer-registry-ubuntu-east-1" {
  value = data.hcp_packer_image.foo.cloud_image_id
}

output "packer-registry-ubuntu-west-1" {
  value = data.hcp_packer_image.bar.cloud_image_id
}
