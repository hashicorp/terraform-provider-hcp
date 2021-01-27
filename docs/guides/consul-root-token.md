---
subcategory: ""
page_title: "Create a new ACL root token - HCP Provider"
description: |-
    An example of creating a new ACL root token.
---

# Create a new Consul ACL root token

Creating a new root token is useful if your HCP Consul cluster has been imported into Terraform
or is managed outside of Terraform. It is important to note that when creating a new root token,
the existing root token will be invalidated.

```terraform
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

// The root_token_accessor_id and root_token_secret_id properties will
// no longer be valid after the new root token is created below
resource "hcp_consul_cluster" "example" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = var.cluster_id
  tier       = "development"
}

// Create a new ACL root token
resource "hcp_consul_cluster_root_token" "example" {
  cluster_id = hcp_consul_cluster.example.cluster_id
}
```
