resource "hcp_vault_secrets_integration_azure" "example" {
  name         = "my-azure-1"
  capabilities = ["ROTATION"]
  client_secret = {
    "tenant_id" : "7eb3...",
    "client_id" : "9de0...",
    "client_secret" : "WZk8..."
  }
}