variable "organization_id" {
  type = string
}

data "hcp_group" "group" {
  resource_name = "iam/organization/${var.organization_id}/group/dev-group"
}

# Note: `roles/vault-radar.resource-viewer` and `roles/vault-radar.resource-contributor` are the only roles
# that can be applied to a policy and/or binding for Vault Radar resources.
resource "hcp_vault_radar_resource_iam_binding" "binding" {
  resource_name = "vault-radar/project/<project_id>/scan-target/<scan_target_id>"
  principal_id  = data.hcp_group.group.resource_id
  role          = "roles/vault-radar.resource-viewer"
}
