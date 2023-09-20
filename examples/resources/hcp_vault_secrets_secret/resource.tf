resource "hcp_vault_secrets_secret" "example" {
  app_name     = "example-app-name"
  secret_name  = "example_secret"
  secret_value = "hashi123"
}