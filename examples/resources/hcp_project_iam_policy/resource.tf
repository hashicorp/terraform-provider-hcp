data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = "roles/contributor"
      principals = [
        "example-user-id-1",
        "example-group-id-1",
        "example-sp-1"
      ]
    },
  ]
}

resource "hcp_project" "my_project" {
  name = "example"
}

resource "hcp_project_iam_policy" "project_policy" {
  project_id  = hcp_project.my_project.resource_id
  policy_data = data.hcp_iam_policy.example.policy_data
}
