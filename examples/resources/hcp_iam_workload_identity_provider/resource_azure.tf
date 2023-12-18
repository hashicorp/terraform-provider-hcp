resource "hcp_service_principal" "workload_sp" {
  name = "my-app-runtime"
}

resource "hcp_iam_workload_identity_provider" "example" {
  name              = "azure-example"
  service_principal = hcp_service_principal.workload_sp.resource_name
  description       = "Allow my-app on Azure to act as my-app-runtime service principal"

  oidc {
    # The issuer uri should be as follows where the ID in the path is replaced
    # with your Azure Tenant ID
    issuer_uri = "https://sts.windows.net/60a0d497-45cd-413d-95ca-e154bbb9129b"

    # The allowed audience should be set to the Object ID of the Azure Managed
    # Identity. In this example, this would be the Object ID of a User Managed
    # Identity that will be attached to "my-app" workloads on Azure.
    allowed_audiences = ["api://10bacc1d-f3f5-499d-a14c-684c1471b27f"]
  }

  # Only allow workload's that are assigned the expected managed identity.
  # The access_token given to Azure workload's will have the oid claim set to
  # that of the managed identity.
  conditional_access = "jwt_claims.oid == `066c643f-86c0-490a-854c-35e77ddc7851`"
}
