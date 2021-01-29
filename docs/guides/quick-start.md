---
subcategory: ""
page_title: "Create resources in HCP - HCP Provider"
description: |-
    An example of creating HCP resources with optional fields defaulted.
---

# Create an HCP HVN, Peering, and Consul cluster

Everything in HashiCorp Cloud Platform (HCP) starts with the HashiCorp Virtual Network (HVN).

HVNs enable you to deploy HashiCorp Cloud products without having to manage the networking details. They give you a simple setup for creating a network on AWS, in the region of your choice, and with the option to specify a CIDR range.

Creating Network peering from your HVN will allow you to connect and launch AWS resources to your HCP account.
Peer your Amazon VPC with your HVN to enable resource access. After creating, you will need to accept the peering request and set up your VPC’s security groups and routing table on your AWS account. The Amazon VPC can be managed with the [AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs).

Once you have an HVN, HCP Consul enables you to quickly deploy Consul servers in AWS across a variety of environments while offloading the operations burden to the SRE experts at HashiCorp.

Finally, with a fully deployed HCP Consul, you need to deploy Consul clients inside of the peered VPC to fully access your Consul features.

```terraform
resource "hcp_hvn" "example_hvn" {
  hvn_id         = "hcp-tf-example-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_consul_cluster" "example_consul_cluster" {
  hvn_id     = hcp_hvn.example_hvn.hvn_id
  cluster_id = "hcp-tf-example-consul-cluster"
  tier       = "development"
}

resource "hcp_aws_network_peering" "example_peering" {
  hvn_id = hcp_hvn.example_hvn.hvn_id

  peer_vpc_id         = "vpc-2f09a348"
  peer_account_id     = "1234567890"
  peer_vpc_region     = "us-west-2"
  peer_vpc_cidr_block = "10.0.1.0/24"
}
```