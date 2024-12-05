variable "github_cloud_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_source_github_cloud" "example" {
  github_organization = "my-github-org"
  token               = var.github_cloud_token
  project_id          = "my-project-id"
}