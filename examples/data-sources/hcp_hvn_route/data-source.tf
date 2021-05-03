data "hcp_hvn_route" "example" {
  hvn              = var.hvn
  destination_cidr = var.destination_cidr
}
