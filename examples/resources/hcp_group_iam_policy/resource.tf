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

# Create an IAM policy for a group using the group manager role
data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = "roles/iam.group-manager"
      principals = [
        data.hcp_user_principal.example.user_id,
      ]
    },
  ]
}

# Set the IAM policy on a group
resource "hcp_group_iam_policy" "example" {
  name        = data.hcp_group.example.resource_name
  policy_data = data.hcp_iam_policy.example.policy_data
}