data "hcp_packer_image_iteration" "alpine" {
  bucket_name = "alpine"
  channel     = "production"
}
