data "hcp_packer_version" "alpine_latest" {
  bucket_name  = "alpine"
  channel_name = "latest"
}

resource "hcp_packer_channel" "alpine_advanced" {
  name        = "advanced"
  bucket_name = data.hcp_packer_version.alpine_latest.bucket_name
}

resource "hcp_packer_channel_assignment" "alpine_advanced" {
  bucket_name         = hcp_packer_channel.alpine_advanced.bucket_name
  channel_name        = hcp_packer_channel.alpine_advanced.name
  version_fingerprint = data.hcp_packer_version.alpine_latest.fingerprint
}
