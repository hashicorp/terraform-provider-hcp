---
page_title: "hcp_aws_transit_gateway_attachment Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The AWS Transit gateway attachment data source provides information about an existing Transit gateway attachment.
---

# Data Source `hcp_aws_transit_gateway_attachment`

The AWS Transit gateway attachment data source provides information about an existing Transit gateway attachment.

## Example Usage

```terraform
data "hcp_aws_transit_gateway_attachment" "test" {
  hvn_id                        = var.hvn_id
  transit_gateway_attachment_id = var.transit_gateway_attachment_id
}
```

## Schema

### Required

- **hvn_id** (String) The ID of the HashiCorp Virtual Network (HVN).
- **transit_gateway_attachment_id** (String) The ID of the Transit gateway attachment.

### Optional

- **id** (String) The ID of this resource.
- **state** (String) The state of the Transit gateway attachment.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **created_at** (String) The time that the Transit gateway attachment was created.
- **destination_cidrs** (List of String) The list of associated CIDR ranges.
- **expires_at** (String) The time after which the Transit gateway attachment will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.
- **organization_id** (String) The ID of the HCP organization where the Transit gateway attachment is located. Always matches the HVN's organization.
- **project_id** (String) The ID of the HCP project where the Transit gateway attachment is located. Always matches the HVN's project.
- **provider_transit_gateway_attachment_id** (String) The Transit gateway attachment ID used by AWS.
- **transit_gateway_id** (String) The ID of the Transit gateway in AWS.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


