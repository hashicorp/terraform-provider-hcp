resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration {
    # Exactly one of `id`, `fingerprint` or `incremental_version` must be passed
    id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
    # fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"
    # incremental_version = 1
  }
}

# To configure a channel to have no assigned iteration, use a "zero value".
# The zero value for `id` and `fingerprint` is `""`; for `incremental_version`, it is `0`
resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration {
    # Exactly one of `id`, `fingerprint` or `incremental_version` must be passed
    id = ""
    # fingerprint = ""
    # incremental_version = 0
  }
}