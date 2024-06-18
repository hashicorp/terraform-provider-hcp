data "hcp_vault_secrets_secret" "secret_data" {
  app_name    = "example-vault-secrets-app"
  secret_name = "my_secret"
}

resource "example_resource" "example" {
  example_attr = data.hcp_vault_secrets_secret.secret_data.secret_value
}

