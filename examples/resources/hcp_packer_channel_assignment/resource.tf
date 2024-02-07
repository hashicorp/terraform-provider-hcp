resource "hcp_packer_channel_assignment" "staging" {
  bucket_name         = "alpine"
  channel_name        = "staging"
  version_fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"
}

# To set the channel to have no assignment, use "none" as the version_fingerprint value.
resource "hcp_packer_channel_assignment" "staging" {
  bucket_name         = "alpine"
  channel_name        = "staging"
  version_fingerprint = "none"
}
