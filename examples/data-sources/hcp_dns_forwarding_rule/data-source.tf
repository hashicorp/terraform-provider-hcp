data "hcp_dns_forwarding_rule" "example" {
  hvn_id                 = var.hvn_id
  dns_forwarding_id      = var.dns_forwarding_id
  dns_forwarding_rule_id = var.dns_forwarding_rule_id
}
