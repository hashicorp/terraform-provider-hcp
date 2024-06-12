# Fetch a user from HCP
data "hcp_user_principal" "example" {
  email = "user@example.com"
}

# Lookup an existing group in HCP
data "hcp_group" "example" {
  resource_name = "group-name"
}

# Add members to the group
resource "hcp_group_members" "example" {
  group = data.hcp_group.example.resource_name
  members = [
    data.hcp_user_principal.example.user_id
  ]
}

# Add an IAM binding to a group
resource "hcp_group_iam_binding" "example" {
  resource_name = data.hcp_group.example.resource_name
  principal_id  = data.hcp_user_principal.example.user_id
  role          = "roles/iam.group-manager"
}