# DNS Forwarding with Transit Gateway Attachment Example
# This example demonstrates how to create DNS forwarding configuration 
# using a Transit Gateway attachment instead of a peering connection.

provider "aws" {
  region = "us-west-2"
}

# HVN for the DNS forwarding
resource "hcp_hvn" "main" {
  hvn_id         = "main-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

# AWS resources for TGW setup
resource "aws_vpc" "example" {
  cidr_block = "172.31.0.0/16"
  tags = {
    Name = "dns-forwarding-vpc"
  }
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "172.31.1.0/24"
}

resource "aws_ec2_transit_gateway" "example" {
  tags = {
    Name = "dns-forwarding-tgw"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}

resource "aws_ram_resource_share" "example" {
  name                      = "dns-forwarding-resource-share"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  principal          = hcp_hvn.main.provider_account_id
}

resource "aws_ram_resource_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  resource_arn       = aws_ec2_transit_gateway.example.arn
}

# HCP Transit Gateway Attachment
resource "hcp_aws_transit_gateway_attachment" "example" {
  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]

  hvn_id                        = hcp_hvn.main.hvn_id
  transit_gateway_attachment_id = "dns-forwarding-tgw-attachment"
  transit_gateway_id            = aws_ec2_transit_gateway.example.id
  resource_share_arn            = aws_ram_resource_share.example.arn
}

# HVN Route for TGW attachment
resource "hcp_hvn_route" "route" {
  hvn_link         = hcp_hvn.main.self_link
  hvn_route_id     = "hvn-to-dns-tgw"
  destination_cidr = aws_vpc.example.cidr_block
  target_link      = hcp_aws_transit_gateway_attachment.example.self_link
}

# DNS Forwarding using TGW attachment
resource "hcp_dns_forwarding" "example_tgw" {
  hvn_id            = hcp_hvn.main.hvn_id
  dns_forwarding_id = "dns-forwarding-tgw"
  peering_id        = hcp_aws_transit_gateway_attachment.example.transit_gateway_attachment_id
  connection_type   = "tgw-attachment"

  forwarding_rule {
    rule_id              = "example-tgw-rule"
    domain_name          = "example-tgw.internal"
    inbound_endpoint_ips = ["172.31.1.10", "172.31.1.11"]
  }
}

# Additional standalone DNS forwarding rule for TGW
resource "hcp_dns_forwarding_rule" "additional_tgw_rule" {
  hvn_id            = hcp_hvn.main.hvn_id
  dns_forwarding_id = hcp_dns_forwarding.example_tgw.dns_forwarding_id
  domain_name       = "api.example-tgw.internal"
  inbound_endpoint_ips = ["172.31.1.12"]
}

# AWS side configuration
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = hcp_aws_transit_gateway_attachment.example.provider_transit_gateway_attachment_id
}

resource "aws_route" "example" {
  route_table_id         = aws_vpc.example.main_route_table_id
  destination_cidr_block = hcp_hvn.main.cidr_block
  transit_gateway_id     = aws_ec2_transit_gateway.example.id
  depends_on = [
    aws_ec2_transit_gateway_vpc_attachment.example
  ]
}

# Data sources for validation
data "hcp_dns_forwarding" "example_tgw" {
  hvn_id            = hcp_hvn.main.hvn_id
  dns_forwarding_id = hcp_dns_forwarding.example_tgw.dns_forwarding_id
}

data "hcp_dns_forwarding_rule" "example_tgw_rule" {
  hvn_id            = hcp_hvn.main.hvn_id
  dns_forwarding_id = hcp_dns_forwarding.example_tgw.dns_forwarding_id
  rule_id           = "example-tgw-rule"
}

# Outputs
output "dns_forwarding_id" {
  description = "The ID of the DNS forwarding configuration"
  value       = hcp_dns_forwarding.example_tgw.id
}

output "tgw_attachment_id" {
  description = "The ID of the Transit Gateway attachment"
  value       = hcp_aws_transit_gateway_attachment.example.transit_gateway_attachment_id
}

output "forwarding_rules" {
  description = "The DNS forwarding rules"
  value = {
    inline_rule     = hcp_dns_forwarding.example_tgw.forwarding_rule
    standalone_rule = hcp_dns_forwarding_rule.additional_tgw_rule.domain_name
  }
}
