resource "hcp_vault_secrets_integration_twilio" "example" {
  name         = "my-twilio-1"
  capabilities = ["ROTATION"]
  static_credential_details = {
    account_sid    = "AC7..."
    api_key_sid    = "TKa..."
    api_key_secret = "6aG..."
  }
}