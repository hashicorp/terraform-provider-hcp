data "hcp_packer_image" "baz" {
  bucket_name    = "hardened-ubuntu-16-04"
  cloud_provider = "aws"
  channel        = "production"
  region         = "us-east-1"
}

output "packer-registry-ubuntu-east-1" {
  value = data.hcp_packer_image.baz.cloud_image_id
}
