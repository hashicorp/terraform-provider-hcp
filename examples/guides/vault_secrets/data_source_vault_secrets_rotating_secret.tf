data "hcp_vault_secrets_rotating_secret" "secret-data" {
  app_name    = "example-vault-secrets-app"
  secret_name = "my_secret"
}

resource "example_resource" "example" {
  example_attr1 = data.hcp_vault_secrets_rotating_secret.secret_values["username"]
  example_attr2 = data.hcp_vault_secrets_rotating_secret.secret_values["password"]
}

