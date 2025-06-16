variable "organization_id" {
  type = string
}

data "hcp_group" "group" {
  resource_name = "iam/organization/${var.organization_id}/group/dev-group"
}

data "hcp_iam_policy" "policy" {
  bindings = [{
    role       = "roles/vault-radar.resource-viewer"
    principals = [hcp_group.group.resource_id]
  }]
}

resource "hcp_vault_radar_resource_iam_policy" "policy" {
  resource_uri = "git://github.com/foo/bar.git"
  policy_data  = data.hcp_iam_policy.policy.policy_data
}
