data "hcp_packer_iteration" "hardened-source" {
  bucket_name = "hardened-ubuntu-16-04"
  channel     = "megan-test"
}