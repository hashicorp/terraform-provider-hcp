---
page_title: "hcp_hvn Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The HVN data source provides information about an existing HashiCorp Virtual Network.
---

# Data Source `hcp_hvn`

The HVN data source provides information about an existing HashiCorp Virtual Network.

## Example Usage

```terraform
data "hcp_hvn" "example" {
  hvn_id         = var.hvn_id
  cloud_provider = var.cloud_provider
  region         = var.region
}
```

## Schema

### Required

- **cloud_provider** (String) The provider where the HVN is located.
- **hvn_id** (String) The ID of the HashiCorp Virtual Network.
- **region** (String) The region where the HVN is located.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **cidr_block** (String) The CIDR range of the HVN.
- **created_at** (String) The time that the HVN was created.
- **organization_id** (String) The ID of the HCP organization where the HVN is located.
- **project_id** (String) The ID of the HCP project where the HVN is located.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


