variable "jira_token" {
  type      = string
  sensitive = true
}

# A Jira subscription requires a Jira connection.
resource "hcp_vault_radar_integration_jira_connection" "jira_connection" {
  name     = "example integration jira connection"
  email    = "jane.smith@example.com"
  token    = var.jira_token
  base_url = "https://example.atlassian.net"
}

resource "hcp_vault_radar_integration_jira_subscription" "jira_subscription" {
  name             = "example integration jira subscription"
  connection_id    = hcp_vault_radar_integration_jira_connection.jira_connection.id
  jira_project_key = "SEC"
  issue_type       = "Task"
  assignee         = "71509:11bb945b-c0de-4bac-9d57-9f09db2f7bc9"
  message          = "Example message"
}