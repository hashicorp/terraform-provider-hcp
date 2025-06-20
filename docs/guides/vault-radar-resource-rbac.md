---
subcategory: ""
page_title: "Managing Vault Radar Resource IAM Policies"
description: |-
    A guide to setting up and managing access to select Radar resources.
---

# Managing Vault Radar Resource IAM Policies

-> **Note:** This feature is currently in private beta.

Administrators can limit users' access to specific Vault Radar resources by using either `hcp_vault_radar_resource_iam_binding` or `hcp_vault_radar_resource_iam_policy`.

Only users with no role at the organization or project level can be restricted to specific Radar resources.

## Pre-requisites and Constraints
* An IAM group should be created with the role `roles/vault-radar.developer` at the project level. This will allow the group's members to access Radar.
* It's recommended to create a group for each team that will be require access to a select set of Radar resources.
* Add users without any roles to the group created above.
* Use the group created above to set the policy or binding for a select set of Radar resources.
* Only the roles `roles/vault-radar.resource-viewer` or `roles/vault-radar.resource-contributor` can be applied to the policy or binding for Radar resources.

## Sample Usage
The following is an example of create a group with the role `roles/vault-radar.developer` at the project level and set the policy for a set of Radar resource that match a resource URI prefix with the role `roles/vault-radar.resource-viewer` for that group.

```terraform
variable "project_id" {
  type = string
}


# Create a group for members with no roles.
resource "hcp_group" "group" {
  display_name = "my-developer-group"
  description  = "my developer group managed by TF"
}

# Assign 'roles/vault-radar.developer' role on the group.
# This allows the groups members access to Vault Radar.
resource "hcp_project_iam_binding" "binding" {
  project_id   = var.project_id
  principal_id = hcp_group.group.resource_id
  role         = "roles/vault-radar.developer"
}

# Create a policy that will grant Radar Resource Viewer access to the group.
data "hcp_iam_policy" "policy" {
  bindings = [
    {
      role       = "roles/vault-radar.resource-viewer"
      principals = [hcp_group.group.resource_id]
    }
  ]
}

# Get the list of Radar resources intended to be accessed by the group.
# This example uses a URI 'LIKE' filter to only include resources that start with "git://github.com/ibm/" or "git://github.com/hashicorp/".
# The % character is a wildcard that matches any sequence of characters.
# Each entry in the uri_like_filter will act like an or condition.
data "hcp_vault_radar_resource_list" "resource_list" {
  uri_like_filter = [
    "git://github.com/ibm/%",
    "git://github.com/hashicorp/%",
  ]
}

# Map the list of Radar resources to a set of HCP resource names, and filter out any resources that are not registered.
locals {
  resources_names = toset(flatten([
    for radar_resource in data.hcp_vault_radar_resource_list.resource_list.resources : radar_resource.hcp_resource_name
    # This is done as a precaution to ensure that only registered resources are processed.
    if radar_resource.hcp_resource_status == "registered"
  ]))
}

# Create IAM policies for each Radar resource's HCP resource name that the group should have access to.
# Note this will replace any existing policies for the resources. If that is not desired, consider using `hcp_vault_radar_resource_iam_binding` instead.
resource "hcp_vault_radar_resource_iam_policy" "policy" {
  for_each      = local.resources_names
  resource_name = each.value
  policy_data   = data.hcp_iam_policy.policy.policy_data
}
```
