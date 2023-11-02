data "hcp_organization" "example_org" {}

resource "hcp_service_principal" "sp" {
  name   = "example-sp"
  parent = data.hcp_organization.example_org.resource_name
}

resource "hcp_organization_iam_binding" "example" {
  principal_id = hcp_service_principal.sp.resource_id
  role         = "roles/contributor"
}
