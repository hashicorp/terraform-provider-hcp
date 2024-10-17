terraform {
  required_providers {
    hcp = {
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}

provider "hcp" {
  client_id = "DPxooi3hyrxklgg1McTzXFFq9WYggZOV"
  client_secret = "5SSM9kj9ccu5gWXDTNE2xVX7dp5XcCXB7L4kPv33vvix6CrYP3Gi6tBkiSAk9dg0"
}
resource "hcp_vault_secrets_integration_confluent" "example" {
  name         = "my-confluent-1"
  capabilities = ["ROTATION"]
  static_credential_details = {
    cloud_api_key_id = "HX33HKM2GCEJMZO7"
    cloud_api_secret = "GRIBCYsIoWp5K6uFY1QWIc09LRmQ8jUI7YGWmQ/wn7UM69LOjyB8EyAVq9DOa9Xb"
  }
}

# resource "hcp_vault_secrets_integration_twilio" "example" {
#   name         = "my-twilio-1"
#   capabilities = ["ROTATION"]
#   static_credential_details = {
#     account_sid    = "AC7..."
#     api_key_sid    = "TKa..."
#     api_key_secret = "6aG..."
#   }
# }