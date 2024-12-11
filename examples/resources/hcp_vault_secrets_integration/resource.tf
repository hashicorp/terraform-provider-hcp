// AWS
resource "hcp_vault_secrets_integration" "example_aws_federated_identity" {
  name          = "my-aws-1"
  capabilities  = ["DYNAMIC", "ROTATION"]
  provider_type = "aws"
  aws_federated_workload_identity = {
    audience = "<audience>>"
    role_arn = "<role-arn>"
  }
}

resource "hcp_vault_secrets_integration" "example_aws_access_keys" {
  name          = "my-aws-2"
  capabilities  = ["DYNAMIC", "ROTATION"]
  provider_type = "aws"
  aws_access_keys = {
    access_key_id     = "<access-key-id>"
    secret_access_key = "<secret-access-key>"
  }
}

// Confluent
resource "hcp_vault_secrets_integration" "example_confluent" {
  name          = "my-confluent-1"
  capabilities  = ["ROTATION"]
  provider_type = "confluent"
  confluent_static_credentials = {
    cloud_api_key_id = "<cloud-api-key-id>"
    cloud_api_secret = "<cloud-api-secret>"
  }
}

// GCP
resource "hcp_vault_secrets_integration" "example_gcp_json_service_account_key" {
  name          = "my-gcp-1"
  capabilities  = ["DYNAMIC", "ROTATION"]
  provider_type = "gcp"
  gcp_service_account_key = {
    credentials = file("${path.module}/my-service-account-key.json")
  }
}

resource "hcp_vault_secrets_integration" "example_gcp_base64_service_account_key" {
  name          = "my-gcp-2"
  capabilities  = ["DYNAMIC", "ROTATION"]
  provider_type = "gcp"
  gcp_service_account_key = {
    credentials = filebase64("${path.module}/my-service-account-key.json")
  }
}

resource "hcp_vault_secrets_integration" "example_gcp_federated_identity" {
  name          = "my-gcp-3"
  capabilities  = ["DYNAMIC", "ROTATION"]
  provider_type = "gcp"
  gcp_federated_workload_identity = {
    service_account_email = "<service-account-email>"
    audience              = "<audience>"
  }
}

// MongoDB-Atlas
resource "hcp_vault_secrets_integration" "example_mongodb_atlas" {
  name          = "my-mongodb-1"
  capabilities  = ["ROTATION"]
  provider_type = "mongodb-atlas"
  mongodb_atlas_static_credentials = {
    api_public_key  = "<api-public-key>"
    api_private_key = "<api-private-key>"
  }
}

// Twilio
resource "hcp_vault_secrets_integration" "example_twilio" {
  name          = "my-twilio-1"
  capabilities  = ["ROTATION"]
  provider_type = "twilio"
  twilio_static_credentials = {
    account_sid    = "<account-sid>"
    api_key_secret = "<api-key-secret>"
    api_key_sid    = "<api-key-sid>"
  }
}