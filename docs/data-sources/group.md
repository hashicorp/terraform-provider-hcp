---
page_title: "hcp_group Data Source - terraform-provider-hcp"
subcategory: "Cloud Platform"
description: |-
  The group data source retrieves the given group.
---

# hcp_group (Data Source)

The group data source retrieves the given group.

## Example Usage

```terraform
data "hcp_group" "example" {
  resource_name = var.resource_name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `resource_name` (String) The group's resource name in format `iam/organization/<organization_id>/group/<group_name>`. The shortened `<group_name>` version can be used for input.

### Read-Only

- `description` (String) The group's description
- `display_name` (String) The group's display name
- `resource_id` (String) The group's unique identifier
