---
page_title: "hcp_vault_cluster_admin_token Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The Vault cluster admin token resource provides a token with administrator privileges on an HCP Vault cluster.
---

# Data Source `hcp_vault_cluster_admin_token`

The Vault cluster admin token resource provides a token with administrator privileges on an HCP Vault cluster.



## Schema

### Required

- **cluster_id** (String) The ID of the HCP Vault cluster.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **token** (String, Sensitive) The admin token of this HCP Vault cluster.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


