data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = "roles/owner"
      principals = [
        "example-user-id-1",
      ]
    },
    {
      role = "roles/admin"
      principals = [
        "example-group-id-1",
        "example-sp-1"
      ]
    },
  ]
}

resource "hcp_organization_iam_policy" "org_policy" {
  policy_data = data.hcp_iam_policy.example.policy_data
}
