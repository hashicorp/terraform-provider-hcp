data "hcp_packer_bucket_names" "all" {}

resource "hcp_packer_channel" "prod" {
  for_each = data.hcp_packer_bucket_names.all.names

  name        = "prod"
  bucket_name = each.key
}
