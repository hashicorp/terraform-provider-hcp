resource "hcp_packer_channel" "advanced" {
  name        = "advanced"
  bucket_name = "alpine"
}

resource "hcp_packer_channel_assignment" "advanced" {
  bucket_name  = hcp_packer_channel.advanced.bucket_name
  channel_name = hcp_packer_channel.advanced.name

  # Exactly one of version, id, or fingerprint must be set:
  iteration_version = 12
  # iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
  # iteration_fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"
}
