resource "hcp_dns_forwarding_rule" "example" {
  hvn_id               = var.hvn_id
  dns_forwarding_id    = var.dns_forwarding_id
  domain_name          = "test.com"
  inbound_endpoint_ips = ["10.0.0.1"]
}
