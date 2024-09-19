resource "hcp_vault_secrets_dynamic_secret" "example_aws" {
  app_name         = "my-app-1"
  secret_provider  = "aws"
  name             = "my_aws_1"
  integration_name = "my-integration-1"
  default_ttl      = "900s"
  aws_assume_role = {
    iam_role_arn = "arn:aws:iam::<account_id>>:role/<role_name>"
  }
}

resource "hcp_vault_secrets_dynamic_secret" "example_gcp" {
  app_name         = "my-app-1"
  secret_provider  = "gcp"
  name             = "my_gcp_1"
  integration_name = "my-integration-1"
  default_ttl      = "900s"
  gcp_impersonate_service_account = {
    service_account_email = "<name>@<project>.iam.gserviceaccount.com"
  }
}