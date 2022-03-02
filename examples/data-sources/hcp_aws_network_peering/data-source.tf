data "hcp_aws_network_peering" "test" {
  hvn_id                = var.hvn_id
  peering_id            = var.peering_id
  wait_for_active_state = true
}
