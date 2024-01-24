data "hcp_packer_version" "hardened-source" {
  bucket_name  = "hardened-ubuntu-16-04"
  channel_name = "production"
}

data "hcp_packer_artifact" "ubuntu-east" {
  bucket_name         = "hardened-ubuntu-16-04"
  version_fingerprint = data.hcp_packer_version.hardened-source.fingerprint
  platform            = "aws"
  region              = "us-east-1"
}

data "hcp_packer_artifact" "ubuntu-west" {
  bucket_name         = "hardened-ubuntu-16-04"
  version_fingerprint = data.hcp_packer_version.hardened-source.fingerprint
  platform            = "aws"
  region              = "us-west-1"
}

output "packer-registry-ubuntu-east-1" {
  value = data.hcp_packer_artifact.ubuntu-east.external_identifier
}

output "packer-registry-ubuntu-west-1" {
  value = data.hcp_packer_artifact.ubuntu-west.external_identifier
}
