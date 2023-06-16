data "hcp_packer_iteration" "alpine_latest" {
  bucket_name = "alpine"
  channel     = "latest"
}

resource "hcp_packer_channel" "alpine_advanced" {
  name        = "advanced"
  bucket_name = data.hcp_packer_iteration.alpine_latest.bucket_name
}

resource "hcp_packer_channel_assignment" "alpine_advanced" {
  bucket_name  = hcp_packer_channel.alpine_advanced.bucket_name
  channel_name = hcp_packer_channel.alpine_advanced.name

  iteration_version = data.hcp_packer_iteration.alpine_latest.incremental_version
}