variable "jira_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_integration_jira_connection" "jira_connection" {
  name     = "example connection to jira"
  email    = "jane.smith@example.com"
  token    = var.jira_token
  base_url = "https://example.atlassian.net"
}