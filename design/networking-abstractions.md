# Networking Abstractions

During early development of the HCP provider, the question of how much abstraction is needed for our networking resources came up. Knowing that our networking resources will be used with various cloud providers, what is the right level of abstraction to use when designing their resource interfaces? What terminology should we use?

We narrowed down to two options:

## 1. Multiple resources, each one specific to a cloud provider

```hcl
resource "hcp_hvn" "example" {
  hvn_id         = "my-favorite-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_aws_peering_connection" "example" {
   hvn_id                = hcp_hvn.example.hvn_id
   peer_vpc_id           = "my-aws-network"
   aws_specific_field    = ...
   ...
}

resource "hcp_gcp_peering_connection" "example" {
   hvn_id                = hcp_hvn.example.hvn_id
   peer_network_id       = "my-google-network"
   gcp_specific_field    = ...
   ...
}

resource "hcp_azure_peering_connection" "example" {
   hvn_id                        = hcp_hvn.example.hvn_id
   virtual_network_peering_name  = "my-azure-network"
   azure_specific_field          = ...
   ...
}
```

## 2. A single resource that accepts a block of cloud provider config

```hcl
resource "hcp_hvn" "example" {
  hvn_id          = "my-favorite-hvn"
  cloud_provider  = "aws"
  region          = "us-west-2"
}

resource "hcp_network_peering" "example" {
  hvn_id               = hcp_hvn.example.hvn_id
    
  aws {
    target_account_id  = "123456789012"
    peer_vpc_id        = "vpc-0054c413d7fb111a1"
    vpc_region         = "us-west-2"
    vpc_cidr_block     = "172.25.8.0/20"
  }
}

resource "hcp_network_peering" "example" {
  hvn_id                  = hcp_hvn.example.hvn_id
    
  gcp { 
    peer_network_id       = "my-google-network"
    gcp_specific_field    = ...
    ...
  }
}
  
resource "hcp_network_peering" "example" {
  hvn_id                          = hcp_hvn.example.hvn_id
    
  azure { 
    virtual_network_peering_name  = "my-azure-network"
    azure_specific_field          = ...
    ... 
   }
}
```

After some user research and discussion with product, the team decided to go with option 1: Multiple resources, each one specific to a cloud provider support by HCP. This decision was made in order to conform with familiar HCL patterns in Terraform today.

It also allows us to tailor each peering resource to its corresponding cloud provider resource. For example, the peering resources in each provider use different words to refer to the network being connected. AWS uses "vpc", Google uses "network", and Azure uses "virtual network". Rather than pick our side in this semantic battle, our networking resources can be designed to match each cloud provider. This reduces the cognitive load of translating between different synonymous terms for users.

Finally, the Terraform Provider SDK does not fully support the block input use case, due to legacy constraints. Separate resources allow us to avoid building a work-around validation system that would be difficult to make obvious through documentation.

Currently, this pattern has been applied to the peering connection and transit gateway attachment resources.

### Exception

There is an exception to this rule: HashiCorp Virtual Networks (HVNs). HVNs and any other networking resource that is designed solely for use within the context of HCP should be represented in the TF provider as a singular cloud provider-agnostic resource that accepts a cloud provider as input.
