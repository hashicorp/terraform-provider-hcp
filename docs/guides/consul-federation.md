---
subcategory: ""
page_title: "Federation with Auto HVN Peering - HCP Provider"
description: |-
    An example of federating a new HCP Consul cluster with an existing one via auto peering.
---

# Federation with Auto HVN Peering

Once you have a HCP Consul cluster, you can create a new Consul cluster to federate with the existing one.
By providing `auto_hvn_to_hvn_peering` as a parameter on the secondary cluster, the HVNs are being peered
automatically ensuring full connectivity. This parameter only ever has to be provided on secondary clusters.

```terraform
resource "hcp_hvn" "primary" {
  hvn_id         = var.primary_hvn_id
  cloud_provider = var.cloud_provider
  region         = var.primary_region
}

resource "hcp_hvn" "secondary" {
  hvn_id         = var.secondary_hvn_id
  cloud_provider = var.cloud_provider
  region         = var.secondary_region
}

resource "hcp_consul_cluster" "primary" {
  hvn_id     = hcp_hvn.primary.hvn_id
  cluster_id = var.primary_cluster_id
  tier       = "development"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id       = hcp_hvn.secondary.hvn_id
  cluster_id   = var.secondary_cluster_id
  tier         = "development"
  primary_link = hcp_consul_cluster.primary.self_link
}
```