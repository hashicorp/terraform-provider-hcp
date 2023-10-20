resource "hcp_service_principal" "workload_sp" {
  name = "my-app-runtime"
}

resource "hcp_iam_workload_identity_provider" "example" {
  service_principal = hcp_service_principal.workload_sp.resource_name
  description       = "Allow my-app on GCP to act as my-app-runtime service principal"

  oidc {
    issuer_uri = "https://accounts.google.com"
  }

  # Only allow workload's that are assigned the expected service account ID
  # GCP will set the subject to that of the service account associated with the
  # workload.
  conditional_access = "jwt_token.sub is `107517467455664443766`"
}
