---
page_title: "hcp_consul_cluster_root_token Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The cluster root token resource is the token used to bootstrap the cluster's ACL system. Using this resource to create a new root token for a cluster will invalidate the consul root token accessor id and Consul root token secret id properties of the cluster.
---

# Resource `hcp_consul_cluster_root_token`

The cluster root token resource is the token used to bootstrap the cluster's ACL system. Using this resource to create a new root token for a cluster will invalidate the consul root token accessor id and Consul root token secret id properties of the cluster.

## Example Usage

```terraform
resource "hcp_consul_cluster_root_token" "example" {
  cluster_id = "consul-cluster"
}
```

## Schema

### Required

- **cluster_id** (String) The ID of the HCP Consul cluster.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **accessor_id** (String) The accessor ID of the root ACL token.
- **kubernetes_secret** (String, Sensitive) The root ACL token Base64 encoded in a Kubernetes secret.
- **secret_id** (String, Sensitive) The secret ID of the root ACL token.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


