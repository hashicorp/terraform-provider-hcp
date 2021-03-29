---
page_title: "hcp_vault_cluster Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The Vault cluster resource allows you to manage an HCP Vault cluster.
---

# Resource `hcp_vault_cluster`

The Vault cluster resource allows you to manage an HCP Vault cluster.



## Schema

### Required

- **cluster_id** (String) The ID of the HCP Vault cluster.
- **hvn_id** (String) The ID of the HVN this HCP Vault cluster is associated to.

### Optional

- **id** (String) The ID of this resource.
- **min_vault_version** (String) The minimum Vault version to use when creating the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.
- **project_id** (String) The ID of the project this HCP Vault cluster is located in.
- **public_endpoint** (Boolean) Denotes that the cluster has a public endpoint for the Vault UI. Defaults to false.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cloud_provider** (String) The provider where the HCP Vault cluster is located.
- **created_at** (String) The time that the Vault cluster was created.
- **namespace** (String) The name of the customer namespace this HCP Vault cluster is located in.
- **organization_id** (String) The ID of the organization this HCP Vault cluster is located in.
- **region** (String) The region where the HCP Vault cluster is located.
- **tier** (String) The tier that the HCP Vault cluster will be provisioned as.  Only 'development' is available at this time.
- **vault_private_endpoint_url** (String) The private URL for the Vault UI.
- **vault_public_endpoint_url** (String) The public URL for the Vault UI. This will be empty if `public_endpoint` is `false`.
- **vault_version** (String) The Vault version of the cluster.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)


