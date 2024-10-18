resource "hcp_vault_secrets_integration_confluent" "example" {
  name         = "my-confluent-1"
  capabilities = ["ROTATION"]
  static_credential_details = {
    cloud_api_key_id = "TKa..."
    cloud_api_secret = "6aG..."
  }
}