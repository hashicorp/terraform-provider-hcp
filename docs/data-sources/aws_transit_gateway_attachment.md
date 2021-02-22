---
page_title: "hcp_aws_transit_gateway_attachment Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The AWS Transit Gateway Attachment data source provides information about an existing transit gateway attachment.
---

# Data Source `hcp_aws_transit_gateway_attachment`

-> **Note:** This feature is currently in private beta. If you would like early access, please [contact our sales team](https://www.hashicorp.com/contact-sales).

The AWS Transit Gateway Attachment data source provides information about an existing transit gateway attachment.

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
- **transit_gateway_attachment_id** (String) The user-settable name of the transit gateway attachment in HCP.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **wait_for_active_state** (Boolean) If `true`, Terraform will wait for the transit gateway attachment to reach an `ACTIVE` state before continuing. Default `false`.

### Read-only

- **created_at** (String) The time that the transit gateway attachment was created.
- **destination_cidrs** (List of String) The list of associated CIDR ranges. Traffic from these CIDRs will be allowed for all resources in the HVN. Traffic to these CIDRs will be routed into this transit gateway attachment.
- **expires_at** (String) The time after which the transit gateway attachment will be considered expired if it hasn't transitioned into `ACCEPTED` or `ACTIVE` state.
- **organization_id** (String) The ID of the HCP organization where the transit gateway attachment is located. Always matches the HVN's organization.
- **project_id** (String) The ID of the HCP project where the transit gateway attachment is located. Always matches the HVN's project.
- **provider_transit_gateway_attachment_id** (String) The transit gateway attachment ID used by AWS.
- **state** (String) The state of the transit gateway attachment.
- **transit_gateway_id** (String) The ID of the user-owned transit gateway in AWS.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)
