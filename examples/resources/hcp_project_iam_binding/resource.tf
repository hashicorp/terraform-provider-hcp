resource "hcp_project" "example" {
  name = "example"
}

resource "hcp_service_principal" "sp" {
  name   = "example-sp"
  parent = hcp_project.example.resource_name
}

resource "hcp_project_iam_binding" "example" {
  project_id   = hcp_project.example.resource_id
  principal_id = hcp_service_principal.sp.resource_id
  role         = "roles/contributor"
}
