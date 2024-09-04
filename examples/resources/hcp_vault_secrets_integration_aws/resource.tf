resource "hcp_vault_secrets_integration_aws" "example_with_access_keys" {
  name         = "my-aws-1"
  capabilities = ["DYNAMIC", "ROTATION"]
  access_keys = {
    access_key_id     = "AKIA..."
    secret_access_key = "rgUK..."
  }
}

resource "hcp_vault_secrets_integration_aws" "example_with_identity_federation" {
  name         = "my-aws-1"
  capabilities = ["DYNAMIC", "ROTATION"]
  federated_workload_identity = {
    role_arn = "arn:aws:iam::<your-account-id>:role/<your-role>>"
    audience = "<your-audience>"
  }
}