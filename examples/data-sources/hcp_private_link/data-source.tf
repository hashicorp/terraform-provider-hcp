data "hcp_private_link" "example" {
  hvn_id          = var.hvn_id
  private_link_id = var.private_link_id
}