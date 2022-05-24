---
subcategory: ""
page_title: "Configure Vault performance replication - HCP Provider"
description: |-
    Configure performance replication between two Plus tier clusters.
---

# Configure Vault performance replication

Admins and Contributors can use the provider to create Plus tier clusters with Vault [performance replication](https://learn.hashicorp.com/tutorials/cloud/vault-replication) enabled via the `primary_link` parameter. In addition to both clusters being in the Plus tier, both must be of the same size (S, M, L).

Although the clusters may reside in the same HVN, it is more likely that you will want to station your performance replication secondary in a different region, and therefore HVN, than your primary. When establishing performance replication links between clusters in different HVNs, an HVN peering connection is required. This can be defined explicitly using an [`hcp_hvn_peering_connection`](../resources/hvn_peering_connection.md), or HCP will create the connection automatically (peering connections can be imported after creation using [terraform import](https://www.terraform.io/cli/import)). Note HVN peering [CIDR block requirements](https://cloud.hashicorp.com/docs/hcp/network/routes#cidr-block-requirements).

-> **Note:** Remember, when scaling performance replicated clusters, be sure to keep the size of all clusters in the group in sync.

### Performance replication example

Clusters configured with performance replication enabled:
```terraform
resource "hcp_hvn" "primary_network" {
  hvn_id         = "hvn1"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_vault_cluster" "primary" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.primary_network.hvn_id
  tier       = "plus_medium"
}

resource "hcp_hvn" "secondary_network" {
  hvn_id         = "hvn2"
  cloud_provider = "aws"
  region         = "eu-central-1"
  cidr_block     = "172.26.16.0/20"
}

resource "hcp_vault_cluster" "secondary" {
  cluster_id   = "vault-cluster"
  hvn_id       = hcp_hvn.secondary_network.hvn_id
  tier         = hcp_vault_cluster.primary.tier
  primary_link = hcp_vault_cluster.primary.self_link
  paths_filter = ["path/a", "path/b"]
}
```
