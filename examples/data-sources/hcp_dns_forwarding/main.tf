terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.5"
    }
  }
}

provider "hcp" {}

data "hcp_dns_forwarding" "example" {
  hvn_id            = "hvn-1"
  dns_forwarding_id = "dns-forwarding-1"
}

output "dns_forwarding_state" {
  value = data.hcp_dns_forwarding.example.state
}

output "forwarding_rules" {
  value = data.hcp_dns_forwarding.example.forwarding_rules
}
