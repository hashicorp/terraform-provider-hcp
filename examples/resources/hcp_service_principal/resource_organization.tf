data "hcp_organization" "my_org" {
}

resource "hcp_service_principal" "example" {
  name   = "example-sp"
  parent = data.hcp_organization.my_org.resource_name
}
