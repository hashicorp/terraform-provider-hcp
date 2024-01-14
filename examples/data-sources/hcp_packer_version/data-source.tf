data "hcp_packer_version" "hardened-source" {
  bucket_name  = "hardened-ubuntu-16-04"
  channel_name = "dev"
}