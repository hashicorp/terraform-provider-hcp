resource "hcp_packer_channel_assignment" "prod" {
  for_each = {
    "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
    "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
  }

  bucket_name = each.key
  # `channel_name` uses a reference to create an implicit dependency so
  # Terraform evaluates the resources in the correct order. This also
  # allows the channel name to be generated dynamically in the
  # `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
