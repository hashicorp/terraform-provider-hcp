resource "hcp_packer_channel_assignment" "prod" {
  for_each = hcp_packer_channel.prod

  bucket_name  = each.key
  channel_name = each.value.name

  iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
}
