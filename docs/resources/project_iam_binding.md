---
page_title: "Resource hcp_project_iam_binding - terraform-provider-hcp"
subcategory: "Cloud Platform"
description: |-
  Updates the project's IAM policy to bind a role to a new member. Existing bindings are preserved.
---

# hcp_project_iam_binding (Resource)

Updates the project's IAM policy to bind a role to a new member. Existing bindings are preserved.

~> **Note:** `hcp_project_iam_binding` can not be used in conjunction with
`hcp_project_iam_policy`.

## Example Usage

```terraform
resource "hcp_project" "example" {
  name = "example"
}

resource "hcp_service_principal" "sp" {
  name   = "example-sp"
  parent = hcp_project.example.resource_name
}

resource "hcp_project_iam_binding" "example" {
  project_id   = hcp_project.example.resource_id
  principal_id = hcp_service_principal.sp.resource_id
  role         = "roles/contributor"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `principal_id` (String) The principal to bind to the given role.
- `role` (String) The role name to bind to the given principal.

### Optional

- `project_id` (String) The ID of the HCP project to apply the IAM Policy to. If unspecified, the project configured on the provider is used.
