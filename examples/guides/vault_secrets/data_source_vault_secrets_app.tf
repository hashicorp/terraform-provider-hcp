data "hcp_vault_secrets_app" "my_app" {
  app_name = "example-vault-secrets-app"
}

resource "example_resource" "example" {
  example_attr = data.hcp_vault_secrets_app.my_app.secrets["my_secret_key"]
}

