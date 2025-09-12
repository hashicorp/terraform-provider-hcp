terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.5"
    }
  }
}

provider "hcp" {}

data "hcp_hvn" "example" {
  hvn_id = "hvn-1"
}

resource "hcp_aws_network_peering" "example" {
  hvn_id          = data.hcp_hvn.example.hvn_id
  peering_id      = "peer-1"
  peer_vpc_id     = "vpc-12345678"
  peer_account_id = "123456789012"
  peer_vpc_region = "us-west-2"
}

resource "hcp_dns_forwarding" "example" {
  hvn_id            = data.hcp_hvn.example.hvn_id
  dns_forwarding_id = "dns-forwarding-1"
  peering_id        = hcp_aws_network_peering.example.peering_id
  connection_type   = "hvn-peering"

  forwarding_rule {
    rule_id              = "rule-1"
    domain_name          = "example.com"
    inbound_endpoint_ips = ["10.0.1.10", "10.0.1.11"]
  }
}
