---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hcp_waypoint_application_template Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The Waypoint Template data source retrieves information on a given Application Template.
---

# hcp_waypoint_application_template (Data Source)

The Waypoint Template data source retrieves information on a given Application Template.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) The ID of the Application Template.
- `name` (String) The name of the Application Template.
- `project_id` (String) The ID of the HCP project where the Waypoint Application Template is located.

### Read-Only

- `description` (String) A description of the template, along with when and why it should be used, up to 500 characters
- `labels` (List of String) List of labels attached to this Application Template.
- `organization_id` (String) The ID of the HCP organization where the Waypoint Application Template is located.
- `readme_markdown_template` (String) Instructions for using the template (markdown format supported)
- `summary` (String) A brief description of the template, up to 110 characters
- `terraform_cloud_workspace_details` (Attributes) Terraform Cloud Workspace details (see [below for nested schema](#nestedatt--terraform_cloud_workspace_details))
- `terraform_no_code_module` (Attributes) Terraform Cloud No-Code Module details (see [below for nested schema](#nestedatt--terraform_no_code_module))

<a id="nestedatt--terraform_cloud_workspace_details"></a>
### Nested Schema for `terraform_cloud_workspace_details`

Read-Only:

- `name` (String) Name of the Terraform Cloud Workspace
- `terraform_project_id` (String) Terraform Cloud Project ID


<a id="nestedatt--terraform_no_code_module"></a>
### Nested Schema for `terraform_no_code_module`

Read-Only:

- `source` (String) No-Code Module Source
- `version` (String) No-Code Module Version
