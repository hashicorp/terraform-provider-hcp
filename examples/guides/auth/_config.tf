terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
    }
  }
}

// Credentials can be set explicitly or via the environment variables HCP_CLIENT_ID and HCP_CLIENT_SECRET
provider "hcp" {
  client_id     = "service-principal-key-client-id"
  client_secret = "service-principal-key-client-secret"
}