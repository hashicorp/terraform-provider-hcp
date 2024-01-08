provider "hcp" {}

provider "aws" {
  region = var.peer_vpc_region
}
