resource "hcp_packer_channel_assignment" "staging" {
  bucket_name  = "alpine"
  channel_name = "staging"

  # Exactly one of version, id, or fingerprint must be set:
  iteration_version = 12
  # iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
  # iteration_fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"
}

# To set the channel to have no assignment, use one of the iteration attributes with their zero value.
# The two string-typed iteration attributes, id and fingerprint, use "none" as their zero value.
resource "hcp_packer_channel_assignment" "staging" {
  bucket_name  = "alpine"
  channel_name = "staging"

  iteration_version = 0
  # iteration_id = "none"
  # iteration_fingerprint = "none"
}