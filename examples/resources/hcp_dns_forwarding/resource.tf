resource "hcp_dns_forwarding" "example" {
  hvn_id             = var.hvn_id
  dns_forwarding_id  = "example-dns-forwarding"
  peering_id         = var.peering_id
  connection_type    = "hvn-peering"
  
  forwarding_rule {
    domain_name           = "example.com"
    inbound_endpoint_ips  = ["10.0.0.1", "10.0.0.2"]
  }
}
