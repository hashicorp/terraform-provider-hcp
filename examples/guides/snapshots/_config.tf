terraform {
  required_providers {
    hcp = {
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}

provider "hcp" {
  client_id     = "service-principal-key-client-id"
  client_secret = "service-principal-key-client-secret"
}