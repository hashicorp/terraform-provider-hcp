terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.5"
    }
  }
}

provider "hcp" {}

data "hcp_dns_forwarding_rule" "example" {
  hvn_id                 = "hvn-1"
  dns_forwarding_id      = "dns-forwarding-1"
  dns_forwarding_rule_id = "rule-1"
}

output "rule_domain" {
  value = data.hcp_dns_forwarding_rule.example.domain_name
}

output "rule_endpoints" {
  value = data.hcp_dns_forwarding_rule.example.inbound_endpoint_ips
}
