resource "hcp_packer_channel_assignment" "staging" {
  bucket_name  = "alpine"
  channel_name = "staging"

  # Exactly one of version, id, or fingerprint must be set:
  iteration_version = 12
  # iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
  # iteration_fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"

  # To set the channel to have no assignment, use one of the attributes with their zero value:
  # iteration_version = 0
  # iteration_id = "none"
  # iteration_fingerprint = "none"
}

# More advanced management is possible, including
# - Creating the channel within Terraform
# - Assigning the channel to the latest complete iteration automatically
data "hcp_packer_iteration" "latest" {
  bucket_name = "alpine"
  channel     = "latest"
}

resource "hcp_packer_channel" "advanced" {
  name        = "advanced"
  bucket_name = data.hcp_packer_iteration.latest.bucket_name
}

resource "hcp_packer_channel_assignment" "advanced" {
  bucket_name  = hcp_packer_channel.advanced.bucket_name
  channel_name = hcp_packer_channel.advanced.name

  iteration_version = data.hcp_packer_iteration.latest.incremental_version
}