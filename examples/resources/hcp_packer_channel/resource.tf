resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
}
