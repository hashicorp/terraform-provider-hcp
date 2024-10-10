variable "slack_token" {
  type      = string
  sensitive = true
}

# A Slack subscription requires a Slack connection.
resource "hcp_vault_radar_integration_slack_connection" "slack_connection" {
  name  = "example connection to slack"
  token = var.slack_token
}

resource "hcp_vault_radar_integration_slack_subscription" "slack_subscription" {
  name          = "example integration slack subscription"
  connection_id = hcp_vault_radar_integration_slack_connection.slack_connection.id
  channel       = "sec-ops-team"
}