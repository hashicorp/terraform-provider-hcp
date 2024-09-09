resource "hcp_vault_secrets_integration_gcp" "example_with_federated_identity" {
  name         = "my-gcp-1"
  capabilities = ["DYNAMIC", "ROTATION"]
  federated_workload_identity = {
    service_account_email = "my-service-account@my-project-123456.iam.gserviceaccount.com"
    audience              = "https://iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/my-identity-pool/providers/my-provider"
  }
}

resource "hcp_vault_secrets_integration_gcp" "example_with_base64_service_account_key" {
  name         = "my-gcp-2"
  capabilities = ["DYNAMIC", "ROTATION"]
  service_account_key = {
    credentials = filebase64("${path.module}/service_account_key.json")
  }
}


resource "hcp_vault_secrets_integration_gcp" "example_json_service_account_key" {
  name         = "my-gcp-3"
  capabilities = ["DYNAMIC", "ROTATION"]
  service_account_key = {
    credentials = file("${path.module}/service_account_key.json")
  }
}