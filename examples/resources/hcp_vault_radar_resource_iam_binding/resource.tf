variable "organization_id" {
  type = string
}

data "hcp_group" "group" {
  resource_name = "iam/organization/${var.organization_id}/group/dev-group"
}

resource "hcp_vault_radar_resource_iam_binding" "binding" {
  resource_uri = "git://github.com/foo/bar.git"
  principal_id = data.hcp_group.group.resource_id
  role         = "roles/vault-radar.resource-viewer"
}
