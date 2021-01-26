---
subcategory: ""
page_title: "Create Consul cluster snapshots in HCP - HCP Provider"
description: |-
    An example of creating an HCP Consul cluster snapshot.
---

# Create Consul cluster snapshots

The snapshot resource allows users to manage Consul snapshots of an HCS cluster. Snapshots currently have a retention policy of 30 days.

```terraform
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

resource "hcp_consul_snapshot" "example" {
  cluster_id    = hcp_consul_cluster.example.cluster_id
  snapshot_name = var.snapshot_name
}
```