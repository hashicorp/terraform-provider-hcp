data "hcp_packer_artifact" "ubuntu-east" {
  bucket_name  = "hardened-ubuntu-16-04"
  channel_name = "production"
  platform     = "aws"
  region       = "us-east-1"
}

output "packer-registry-ubuntu-east-1" {
  value = data.hcp_packer_artifact.ubuntu-east.external_identifier
}
