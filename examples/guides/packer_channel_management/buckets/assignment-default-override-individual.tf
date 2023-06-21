resource "hcp_packer_channel_assignment" "prod" {
  for_each = hcp_packer_channel.prod

  bucket_name  = each.key
  channel_name = each.value.name

  iteration_id = lookup(
    { # Per-bucket override for the iteration ID
      "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
      "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
    },
    each.key,
    "none"
  )
}
