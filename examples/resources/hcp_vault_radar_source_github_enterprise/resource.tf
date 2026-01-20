variable "github_enterprise_token" {
  type      = string
  sensitive = true
}

# Example 1: Basic GitHub Enterprise source (HCP detector mode - default)
# Minimal configuration - detector_type defaults to 'hcp' when not specified
resource "hcp_vault_radar_source_github_enterprise" "example_basic" {
  domain_name         = "myserver.acme.com"
  github_organization = "my-github-org"
  token               = var.github_enterprise_token
}

# Example 2: GitHub Enterprise source with HCP detector and secret copying enabled
# Specifies token_env_var to enable secret copying via Vault Radar Agent
resource "hcp_vault_radar_source_github_enterprise" "example_hcp_with_secret_copying" {
  domain_name         = "myserver.acme.com"
  github_organization = "my-github-org"
  token               = var.github_enterprise_token
  token_env_var       = "GITHUB_ENTERPRISE_TOKEN" # Enables secret copying via Vault Radar Agent
  detector_type       = "hcp"
}

# Example 3: GitHub Enterprise source with agent detector
# Agent mode requires token_env_var and forbids token
resource "hcp_vault_radar_source_github_enterprise" "example_agent" {
  domain_name         = "myserver.acme.com"
  github_organization = "my-github-org"
  token_env_var       = "GITHUB_ENTERPRISE_TOKEN"
  detector_type       = "agent"
}

