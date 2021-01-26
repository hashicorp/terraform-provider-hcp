---
page_title: "hcp_consul_snapshot Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The Consul snapshot resource allows users to manage Consul snapshots of an HCP Consul cluster. Snapshots currently have a retention policy of 30 days.
---

# Resource `hcp_consul_snapshot`

The Consul snapshot resource allows users to manage Consul snapshots of an HCP Consul cluster. Snapshots currently have a retention policy of 30 days.



## Schema

### Required

- **cluster_id** (String) The ID of the HCP Consul cluster.
- **snapshot_name** (String) The name of the snapshot.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **consul_version** (String) The version of Consul at the time of snapshot creation.
- **project_id** (String) The ID of the project the HCP Consul cluster is located.
- **restored_at** (String) Timestamp of when the snapshot was restored. If the snapshot has not been restored, this field will be blank.
- **size** (Number) The size of the snapshot in bytes.
- **snapshot_id** (String) The ID of the Consul snapshot

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)
- **update** (String)


