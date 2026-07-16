data "hcp_organization" "example" {}

resource "hcp_resource_control_policy" "example" {
  organization_id = data.hcp_organization.example.resource_id

  enabled_constraints = [
    "constraints/packer.deny-creation",
    "constraints/vault.deny-creation",
  ]
}
