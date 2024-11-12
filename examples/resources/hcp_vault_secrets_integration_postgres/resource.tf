resource "hcp_vault_secrets_integration_postgres" "example" {
  name         = "my-postgres-1"
  capabilities = ["ROTATION"]
  static_credential_details = {
    connection_string = "postgres://user:password@localhost:5432/dbname"
  }
}
