data "hcp_aws_transit_gateway_attachment" "test" {
  hvn_id                        = var.hvn_id
  transit_gateway_attachment_id = var.transit_gateway_attachment_id
}
