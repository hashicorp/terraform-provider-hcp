resource "hcp_service_principal" "example" {
  name = "example-sp"
}

resource "hcp_service_principal_key" "key" {
  service_principal = hcp_service_principal.example.resource_name
}
