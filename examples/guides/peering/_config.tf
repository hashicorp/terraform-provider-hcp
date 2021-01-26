terraform {
  required_providers {
    hcp = {
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}
provider "hcp" {}