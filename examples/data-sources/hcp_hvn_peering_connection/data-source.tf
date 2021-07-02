data "hcp_hvn_peering_connection" "test" {
  peering_id = var.peering_id
  hvn_1      = var.hvn_1
  hvn_2      = var.hvn_2
}
