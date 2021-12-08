---
subcategory: ""
page_title: "Resize or scale a Vault cluster - HCP Provider"
description: |-
    Change a current HCP Vault cluster in terms of tiers (Dev, Starter, Standard) or sizes (S, M, L).
---

# Scale a cluster

Admins are able to use the provider to change a cluster’s size or tier. Scaling down to a Development tier from any production-grade tier is not allowed. In addition, if you are using too much storage and want to scale down to a smaller size or tier, you will be unable to do so until you delete enough resources. 

### Scaling example

Initial Cluster:
```terraform
resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_vault_cluster" "example" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.example.hvn_id
  # default tier is “dev”
}
```

Scaling to Standard:
```terraform
resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_vault_cluster" "example" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.example.hvn_id
  tier       = "standard_medium"
}
```

