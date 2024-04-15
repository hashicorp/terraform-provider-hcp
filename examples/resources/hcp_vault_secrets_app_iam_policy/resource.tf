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


resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_policy" "example" {
  resource_name = "secrets/project/41d107a7-eea6-4b5e-8481-508ab29e2b07/app/example-app-name"
  policy_data   = data.hcp_iam_policy.example.policy_data
}
