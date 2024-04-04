data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = "roles/admin"
      principals = [
        "example-user-id-1",
        "example-group-id-1",
        "example-sp-1"
      ]
    },
    {
      role = "roles/contributor"
      principals = [
        "example-user-id-2",
        "example-group-id-2",
      ]
    },
    {
      role = "roles/secrets.app-secret-reader"
      principals = [
        "example-sp-3"
      ]
    },
  ]
}
