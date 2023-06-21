resource "hcp_packer_channel_assignment" "prod" {
  for_each = merge(
    { for c in hcp_packer_channel.prod : c.bucket_name => "none" },
    # If a default value is not desired, omit the line above, and buckets that 
    # aren't added to the map won't have a channel assignment set.
    # If a default value is not desired, but all channels should have an 
    # assignment set, replace `"none"` with `null` to ensure that every channel
    # is covered by at least one of the filters.
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H1SF9NWAK8AP25PAWDBGZ1YD"
      if startswith(v.bucket_name, "prefix1")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H28NK6V40TKSC4MMD3Z5NGMN"
      if startswith(v.bucket_name, "prefix2")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H28NJ7WPCZA0FZZ8G3FGGTAF"
      if endswith(v.bucket_name, "someSuffix")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H1SF9NWAK8AP25PAWDBGZ1YD"
      if strcontains(v.bucket_name, "someContents")
    },
  )

  bucket_name = each.key
  # Using a reference for `channel_name` allows it to be generated dynamically
  # in the `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
