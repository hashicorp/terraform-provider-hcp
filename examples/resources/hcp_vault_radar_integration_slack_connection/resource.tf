variable "slack_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_integration_slack_connection" "slack_connection" {
  name  = "example connection to slack"
  token = var.slack_token
}
