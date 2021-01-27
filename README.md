# HashiCorp Cloud Platform (HCP) Terraform Provider

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
-	[Go](https://golang.org/doc/install) >= 1.14

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make dev` command: 
```sh
$ make dev
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

## Generating Docs

To generate or update documentation, run `go generate`.
```shell script
$ go generate
```

## Using the provider

Please see the docs for details about a particular resource. 
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
  region         = "us-west-1"
  cidr_block     = "172.25.16.0/20"
}

// Create an HCP Consul cluster within the HVN.
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = "consul-cluster"
  cloud_provider = hcp_hvn.example.cloud_provider
  region         = hcp_hvn.example.region
}

// If you have not already, create a VPC within your AWS account that will
// contain the workloads you want to connect to your HCP Consul cluster.
// Make sure the CIDR block of the peer VPC does not overlap with the CIDR
// of the HVN.
resource "aws_vpc" "peer" {
  cidr_block = "10.220.0.0/16"
}

// Create an HCP Peering Connection to peer your HVN with your AWS VPC.
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