resource "hcp_vault_secrets_rotating_secret" "example_aws" {
  app_name             = "my-app-1"
  secret_provider      = "aws"
  name                 = "my_aws_1"
  integration_name     = "my-aws-1"
  rotation_policy_name = "built-in:60-days-2-active"
  aws_access_keys = {
    iam_username = "my-iam-username"
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_gcp" {
  app_name             = "my-app-1"
  secret_provider      = "gcp"
  name                 = "my_gcp_1"
  integration_name     = "my-gcp-1"
  rotation_policy_name = "built-in:60-days-2-active"
  gcp_service_account_key = {
    service_account_email = "<name>>@<project>.iam.gserviceaccount.com"
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_mongodb_atlas" {
  app_name             = "my-app-1"
  secret_provider      = "mongodb_atlas"
  name                 = "my_mongodb_atlas_1"
  integration_name     = "my-mongodbatlas-1"
  rotation_policy_name = "built-in:60-days-2-active"
  mongodb_atlas_user = {
    project_id    = "<uuid>>"
    database_name = "my-cluster-0"
    roles         = ["readWrite", "read"]
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_twilio" {
  app_name             = "my-app-1"
  secret_provider      = "twilio"
  name                 = "my_twilio_1"
  integration_name     = "my-twilio-1"
  rotation_policy_name = "built-in:60-days-2-active"
  twilio_api_key       = {}
}

resource "hcp_vault_secrets_rotating_secret" "example_confluent" {
  app_name             = "my-app-1"
  secret_provider      = "confluent"
  name                 = "my_confluent_1"
  integration_name     = "my-confluent-1"
  rotation_policy_name = "built-in:60-days-2-active"
  confluent_service_account = {
    service_account_id = "<service-account-id>"
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_postgres" {
  app_name             = "my-app-1"
  secret_provider      = "confluent"
  name                 = "my_confluent_1"
  integration_name     = "my-confluent-1"
  rotation_policy_name = "built-in:60-days-2-active"
  postgres_usernames = {
    usernames = ["user1", "user2"]
  }
}

