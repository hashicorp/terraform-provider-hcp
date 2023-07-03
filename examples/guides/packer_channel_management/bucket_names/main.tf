data "hcp_packer_bucket_names" "all" {}

resource "hcp_packer_channel" "release" {
  for_each = data.hcp_packer_bucket_names.all.names

  name        = "release"
  bucket_name = each.key
}
