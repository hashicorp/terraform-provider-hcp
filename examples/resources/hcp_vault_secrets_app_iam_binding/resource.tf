resource "hcp_service_principal" "sp" {
  name = "example-sp"
}

resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_binding" "example" {
  resource_name = hcp_vault_secrets_app.example.resource_name
  principal_id  = hcp_service_principal.sp.resource_id
  role          = "roles/secrets.app-secret-reader"
}
