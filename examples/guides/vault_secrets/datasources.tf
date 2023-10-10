data "hcp_vault_secrets_app" "app-data" {
  app_name = "example-vault-secrets-app"
}
resource "example_resource" "example" {
  example_attr = data.hcp_vault_secrets_secret.secret_one
}