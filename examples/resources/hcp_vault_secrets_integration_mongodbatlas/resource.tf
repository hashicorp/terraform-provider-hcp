resource "hcp_vault_secrets_integration_mongodbatlas" "example" {
  name         = "my-mongodbatlas-1"
  capabilities = ["ROTATION"]
  static_credential_details = {
    api_private_key = "12345678-1234-1234-1234-12345678"
    api_public_key = "abcdefgh"
  }
}