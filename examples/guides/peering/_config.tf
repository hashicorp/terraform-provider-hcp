terraform {
  required_providers {
    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.1.0"
    }
    aws = "~> 2.64.0"
  }
}

provider "hcp" {}

provider "aws" {
  region = "us-east-1"
}