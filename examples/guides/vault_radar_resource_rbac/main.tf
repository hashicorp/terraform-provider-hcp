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

# Map the list of Radar resources to a map of uris to HCP resource names, and filter out any resources that are not registered.
locals {
  resources_uri_to_resource_name = {
    for radar_resource in data.hcp_vault_radar_resource_list.resource_list.resources : radar_resource.uri =>
    radar_resource.hcp_resource_name
    # This is done as a precaution to ensure that only valid resources are processed.
    if radar_resource.hcp_resource_status == "registered"
  }

}

# Create IAM policies for each Radar resource's HCP resource name that the group should have access to.
# Note this will replace any existing policies for the resources. If that is not desired, consider using `hcp_vault_radar_resource_iam_binding` instead.
resource "hcp_vault_radar_resource_iam_policy" "policy" {
  for_each      = local.resources_uri_to_resource_name
  resource_name = each.value
  policy_data   = data.hcp_iam_policy.policy.policy_data
}

