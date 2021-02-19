---
subcategory: ""
page_title: "Federate HCP Consul clusters - HCP Provider"
description: |-
    An example of federating a new HCP Consul cluster with an existing one.
---

# Federate a new HCP Consul cluster with an existing one

Once you have a HCP Consul cluster, you can create a new Consul cluster to federate with the existing one.

```terraform
resource "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}

resource "hcp_consul_cluster" "primary" {
  hvn_id     = hcp_hvn.example.hvn_id
  cluster_id = var.primary_cluster_id
  tier       = "development"
}

resource "hcp_consul_cluster" "secondary" {
  hvn_id       = hcp_hvn.example.hvn_id
  cluster_id   = var.secondary_cluster_id
  tier         = "development"
  primary_link = hcp_consul_cluster.primary.self_link
}
```