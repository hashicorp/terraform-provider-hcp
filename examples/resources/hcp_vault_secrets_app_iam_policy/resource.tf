data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = "roles/secrets.app-secret-reader"
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
  resource_name = hcp_vault_secrets_app.example.resource_name
  policy_data   = data.hcp_iam_policy.example.policy_data
}
