resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}
resource "hcp_vault_secrets_secret" "secret-example" {
  app_name     = hcp.hcp_vault_secrets_app.example.app_name
  secret_name  = "a-new-secret"
  secret_value = "a test secret"
}