resource "hcp_packer_channel_assignment" "prod" {
  for_each = merge(
    { for c in hcp_packer_channel.prod : c.bucket_name => "none" },
    {
      "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
      "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
    }
  )

  bucket_name = each.key
  # Using a reference for `channel_name` allows it to be generated dynamically
  # in the `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
