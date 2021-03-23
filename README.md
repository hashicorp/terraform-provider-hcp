<p align="center" style="text-align:center;">
  <img alt="HashiCorp Cloud Platform logo" src="hcp.svg" width="300" />
</p>

# HashiCorp Cloud Platform (HCP) Terraform Provider

Available in the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/hcp/latest).

The HashiCorp Cloud Platform (HCP) Terraform Provider is a plugin for Terraform that allows for the full lifecycle management of HCP resources. This provider is maintained internally by the HashiCorp Cloud Services team.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x

## Using the Provider

See the [HashiCorp Cloud Platform (HCP) Provider documentation](https://registry.terraform.io/providers/hashicorp/hcp/latest/docs) to get started using the provider.

## Contributing

See the [`contributing`](contributing/) directory for more developer documentation.

## Example

Below is a complex example that creates a HashiCorp Virtual Network (HVN), an HCP Consul cluster within that HVN, and peers the HVN to an AWS VPC.
```hcl
// Configure the provider
provider "hcp" {}

provider "aws" {
  region = "us-west-2"
}

// Create a HashiCorp Virtual Network (HVN).
resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

// Create an HCP Consul cluster within the HVN.
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = "consul-cluster"
  tier           = "development"
}

// If you have not already, create a VPC within your AWS account that will
// contain the workloads you want to connect to your HCP Consul cluster.
// Make sure the CIDR block of the peer VPC does not overlap with the CIDR
// of the HVN.
resource "aws_vpc" "peer" {
  cidr_block = "10.220.0.0/16"
}

// Create an HCP network peering to peer your HVN with your AWS VPC.
resource "hcp_aws_network_peering" "example" {
  peering_id          = "peer-id"
  hvn_id              = hcp_hvn.example.hvn_id
  peer_vpc_id         = aws_vpc.peer.id
  peer_account_id     = aws_vpc.peer.owner_id
  peer_vpc_region     = "us-west-2"
  peer_vpc_cidr_block = aws_vpc.peer.cidr_block
}

// Accept the VPC peering within your AWS account.
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = hcp_aws_network_peering.example.provider_peering_id
  auto_accept               = true
}
```
