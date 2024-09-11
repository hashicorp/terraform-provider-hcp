variable "github_enterprise_token" {
  type = string
}

resource "hcp_vault_radar_source_github_enterprise" "example" {
  domain_name         = "myserver.acme.com"
  github_organization = "my-github-org"
  token               = var.github_enterprise_token
  project_id          = "my-project-id"
}