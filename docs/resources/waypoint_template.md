---
page_title: "hcp_waypoint_template Resource - terraform-provider-hcp"
subcategory: "HCP Waypoint"
description: |-
  Waypoint Template resource
---

# hcp_waypoint_template `Resource`



Waypoint Template resource

## Example Usage

```terraform
resource "tfe_project" "example" {
  name         = "waypoint-build-destination"
  organization = var.org_name
}

data "tfe_registry_module" "example" {
  organization    = var.org_name
  name            = "my-nocode-example-module"
  module_provider = "aws"
}

resource "hcp_waypoint_template" "example" {
  name                            = "example-aws-template"
  summary                         = "AWS waypoint deployment."
  description                     = "Deploys a nocode module."
  terraform_project_id            = tfe_project.example.id
  labels                          = ["pets"]
  terraform_no_code_module_source = data.tfe_registry_module.example.no_code_module_source
  terraform_no_code_module_id     = data.tfe_registry_module.example.no_code_module_id
  variable_options = [
    {
      name          = "resource_size"
      user_editable = true
      options       = ["small", "medium", "large"]
    },
    {
      name          = "service_port"
      user_editable = false
      options       = ["8080"]
    },
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Template.
- `summary` (String) A brief description of the template, up to 110 characters.
- `terraform_no_code_module_id` (String) The ID of the Terraform no-code module to use for running Terraform operations. This is in the format of 'nocode-<ID>'.
- `terraform_no_code_module_source` (String) Terraform Cloud No-Code Module details
- `terraform_project_id` (String) The ID of the Terraform Cloud Project to create workspaces in. The ID is found on the Terraform Cloud Project settings page.

### Optional

- `actions` (List of String) List of actions by 'ID' to assign to this Template. Applications created from this Template will have these actions assigned to them. Only 'ID' is supported.
- `description` (String) A description of the template, along with when and why it should be used, up to 500 characters
- `labels` (List of String) List of labels attached to this Template.
- `project_id` (String) The ID of the HCP project where the Waypoint Template is located.
- `readme_markdown_template` (String) Instructions for using the template (markdown format supported).
- `terraform_agent_pool_id` (String) The ID of the agent pool to use for Terraform operations, for workspaces created for applications using this template. Required if terraform_execution_mode is set to 'agent'.
- `terraform_cloud_workspace_details` (Attributes, Deprecated) Terraform Cloud Workspace details (see [below for nested schema](#nestedatt--terraform_cloud_workspace_details))
- `terraform_execution_mode` (String) The execution mode of the HCP Terraform workspaces created for applications using this template.
- `use_module_readme` (Boolean) If true, will auto-import the readme form the Terraform module used. If this is set to true, users should not also set `readme_markdown_template`.
- `variable_options` (Attributes Set) List of variable options for the template. (see [below for nested schema](#nestedatt--variable_options))

### Read-Only

- `id` (String) The ID of the Template.
- `organization_id` (String) The ID of the HCP organization where the Waypoint Template is located.

<a id="nestedatt--terraform_cloud_workspace_details"></a>
### Nested Schema for `terraform_cloud_workspace_details`

Required:

- `name` (String) Name of the Terraform Cloud Project
- `terraform_project_id` (String) Terraform Cloud Project ID


<a id="nestedatt--variable_options"></a>
### Nested Schema for `variable_options`

Required:

- `name` (String) Variable name
- `variable_type` (String) Variable type

Optional:

- `options` (List of String) List of options
- `user_editable` (Boolean) Whether the variable is editable by the user creating an application
