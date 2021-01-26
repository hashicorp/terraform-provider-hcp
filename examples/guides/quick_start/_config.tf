terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.1.0"
    }
  }
}

provider "hcp" {}
