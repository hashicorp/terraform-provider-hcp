---
subcategory: ""
page_title: "Create Consul cluster snapshots in HCP - HCP Provider"
description: |-
    An example of creating an HCP Consul cluster snapshot.
---

# Create Consul cluster snapshots

TODO: summarize creating snapshots

```terraform
# add comment
resource "hcp_consul_cluster" "example" {
  hvn_id         = hcp_hvn.example.hvn_id
  cluster_id     = var.cluster_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

# TODO snapshots
```