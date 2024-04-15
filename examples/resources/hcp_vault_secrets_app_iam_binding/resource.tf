resource "hcp_service_principal" "sp" {
  name = "example-sp"
}

resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_binding" "example" {
  resource_name = "secrets/project/41d107a7-eea6-4b5e-8481-508ab29e2b07/app/example-app-name"
  principal_id  = hcp_service_principal.sp.resource_id
  role          = "roles/viewer"
}
