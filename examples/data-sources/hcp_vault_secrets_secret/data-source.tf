data "hcp_vault_secrets_secret" "example" {
  app_name    = "example-vault-secrets-app"
  secret_name = "my_secret"
}
