---
page_title: "hcp_hvn Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The HVN resource allows you to manage a HashiCorp Virtual Network in HCP.
---

# Resource `hcp_hvn`

The HVN resource allows you to manage a HashiCorp Virtual Network in HCP.

## Example Usage

```terraform
resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cidr_block     = "172.25.16.0/20"
  cloud_provider = "aws"
  region         = "us-west-2"
  project_id     = var.project_id
}
```

## Schema

### Required

- **cloud_provider** (String) The provider where the HVN is located. Only 'aws' is available at this time.
- **hvn_id** (String) The ID of the HashiCorp Virtual Network.
- **project_id** (String) The ID of the HCP project where the HVN is located.
- **region** (String) The region where the HVN is located.

### Optional

- **cidr_block** (String) The CIDR range of the HVN.
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)


