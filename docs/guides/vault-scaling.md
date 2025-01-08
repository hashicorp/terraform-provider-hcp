---
subcategory: ""
page_title: "Resize or scale a Vault cluster"
description: |-
    Change a current HCP Vault cluster in terms of tiers (Dev, Starter, Standard, Plus) or sizes (S, M, L).
---

# Scale a cluster

Admins are able to use the provider to change a cluster’s size or tier. There are a few limitations on cluster scaling:

- When scaling performance replicated Plus-tier clusters, be sure to keep the size of all clusters in the group in sync
- Scaling down to the Development tier from any production-grade tier is not allowed
- If you are using too much storage and want to scale down to a smaller size or tier, you will be unable to do so until you delete enough resources

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

