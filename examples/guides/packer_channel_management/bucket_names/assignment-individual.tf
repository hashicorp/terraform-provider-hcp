resource "hcp_packer_channel_assignment" "bucket1prod" {
  bucket_name = "bucket1"
  # `channel_name` uses a reference to create an implicit dependency so
  # Terraform evaluates the resources in the correct order. This also
  # allows the channel name to be generated dynamically in the
  # `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod["bucket1"].name

  iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
}

resource "hcp_packer_channel_assignment" "bucket2prod" {
  bucket_name  = "bucket2"
  channel_name = hcp_packer_channel.prod["bucket2"].name

  iteration_fingerprint = "01H1GH3NWAK8AN98QWDSALN3VO"
}