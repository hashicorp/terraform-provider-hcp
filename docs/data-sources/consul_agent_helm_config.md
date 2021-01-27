---
page_title: "hcp_consul_agent_helm_config Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The Consul agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.
---

# Data Source `hcp_consul_agent_helm_config`

The Consul agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.

## Example Usage

```terraform
data "hcp_consul_agent_helm_config" "example" {
  cluster_id          = var.cluster_id
  kubernetes_endpoint = var.kubernetes_endpoint
}
```

## Schema

### Required

- **cluster_id** (String) The ID of the HCP Consul cluster.
- **kubernetes_endpoint** (String) The FQDN for the Kubernetes API.

### Optional

- **expose_gossip_ports** (Boolean) Denotes that the gossip ports should be exposed.
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **config** (String) The agent Helm config.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


