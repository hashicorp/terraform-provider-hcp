data "hcp_packer_image_iteration" "alpine" {
  bucket  = "alpine"
  channel = "production"
}
