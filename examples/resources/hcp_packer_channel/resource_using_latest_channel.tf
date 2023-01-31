data "hcp_packer_image_iteration" "latest" {
  bucket_name = "alpine"
  channel     = "latest"
}

resource "hcp_packer_channel" "staging" {
  name        = staging
  bucket_name = alpine
  iteration {
    id = data.hcp_packer_image_iteration.latest.id
  }
}
