---
page_title: "hcp_consul_cluster Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The Consul cluster resource allows you to manage an HCP Consul cluster.
---

# Resource `hcp_consul_cluster`

The Consul cluster resource allows you to manage an HCP Consul cluster.



## Schema

### Required

- **cloud_provider** (String) The provider where the HCP Consul cluster is located. Only 'aws' is available at this time.
- **cluster_id** (String) The ID of the HCP Consul cluster.
- **hvn_id** (String) The ID of the HVN this HCP Consul cluster is associated to.
- **region** (String) The region where the HCP Consul cluster is located.

### Optional

- **connect_enabled** (Boolean) Denotes the Consul connect feature should be enabled for this cluster.  Default to true.
- **datacenter** (String) The Consul data center name of the cluster. If not specified, it is defaulted to the value of `cluster_id`.
- **id** (String) The ID of this resource.
- **min_consul_version** (String) The minimum Consul version of the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.
- **public_endpoint** (Boolean) Denotes that the cluster has a public endpoint for the Consul UI. Defaults to false.
- **tier** (String) The feature tier that the HCP Consul cluster will be provisioned as.  Only 'dev' and 'standard' are available at this time.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **consul_automatic_upgrades** (Boolean) Denotes that automatic Consul upgrades are enabled.
- **consul_ca_file** (String) The cluster CA file encoded as a Base64 string.
- **consul_config_file** (String) The cluster config encoded as a Base64 string.
- **consul_private_endpoint_url** (String) The private URL for the Consul UI.
- **consul_public_endpoint_url** (String) The public URL for the Consul UI. This will be empty if `public_endpoint` is `false`.
- **consul_root_token_accessor_id** (String) The accessor ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.
- **consul_root_token_secret_id** (String, Sensitive) The secret ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.
- **consul_snapshot_interval** (String) The Consul snapshot interval.
- **consul_snapshot_retention** (String) The retention policy for Consul snapshots.
- **consul_version** (String) The Consul version of the cluster.
- **num_servers** (Number) The the number of Consul server nodes in the cluster.
- **organization_id** (String) The ID of the organization this HCP Consul cluster is located in.
- **project_id** (String) The ID of the project this HCP Consul cluster is located in.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)
- **update** (String)


