# the provider is derived from the integration name
resource "hcp_vault_secrets_sync" "example_aws_sync" {
  name             = "my-aws-1"
  integration_name = "my-integration-1"
}
