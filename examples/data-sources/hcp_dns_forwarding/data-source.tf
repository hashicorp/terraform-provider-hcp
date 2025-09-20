data "hcp_dns_forwarding" "example" {
  hvn_id            = var.hvn_id
  dns_forwarding_id = var.dns_forwarding_id
}
