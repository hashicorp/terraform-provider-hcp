---
page_title: "hcp_consul_cluster_root_token Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The Consul cluster resource allows you to manage an HCP Consul cluster.
---

# Resource `hcp_consul_cluster_root_token`

The Consul cluster resource allows you to manage an HCP Consul cluster.

## Example Usage

```terraform
resource "hcp_consul_cluster_root_token" "example" {
  cluster_id = var.cluster_id
  project_id = var.project_id
}
```

## Schema

### Required

- **cluster_id** (String) The ID of the HCP Consul cluster.

### Optional

- **id** (String) The ID of this resource.
- **project_id** (String) The ID of the project this HCP Consul cluster is located.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **accessor_id** (String) The accessor ID of the root ACL token.
- **kubernetes_secret** (String, Sensitive) The root ACL token Base64 encoded in a Kubernetes secret.
- **secret_id** (String, Sensitive) The secret ID of the root ACL token.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


