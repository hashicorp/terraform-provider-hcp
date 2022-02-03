data "hcp_azure_peering_connection" "test" {
  hvn_id     = var.hvn_id
  peering_id = var.peering_id
}
