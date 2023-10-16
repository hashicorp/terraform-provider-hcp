resource "hcp_project" "my_proj" {
  name = "example"
}

resource "hcp_service_principal" "example" {
  name   = "example-sp"
  parent = hcp_project.my_proj.resource_name
}
