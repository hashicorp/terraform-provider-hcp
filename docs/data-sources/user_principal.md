---
page_title: "hcp_user_principal Data Source - terraform-provider-hcp"
subcategory: "Cloud Platform"
description: |-
  The user principal data source retrieves the given user principal.
---

# hcp_user_principal (Data Source)

The user principal data source retrieves the given user principal.

## Example Usage

```terraform
data "hcp_user_principal" "example" {
  user_id = var.example_user_id
}

data "hcp_user_principal" "example" {
  email = var.example_email
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `email` (String) The user's email. Can not be combined with user_id.
- `user_id` (String) The user's unique identifier. Can not be combined with email.
