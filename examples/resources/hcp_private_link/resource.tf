provider "aws" {
  region = "us-west-2"
}

# Create an HVN
resource "hcp_hvn" "example" {
  hvn_id         = "private-link-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

# Create a Vault cluster
resource "hcp_vault_cluster" "example" {
  cluster_id      = "vault-cluster"
  hvn_id          = hcp_hvn.example.hvn_id
  tier            = "standard_small"
  public_endpoint = false

  major_version_upgrade_config {
    upgrade_type = "AUTOMATIC"
  }
}

# Create a Private Link for the Vault cluster
resource "hcp_private_link" "example" {
  hvn_id           = hcp_hvn.example.hvn_id
  private_link_id  = "example-private-link"
  vault_cluster_id = hcp_vault_cluster.example.cluster_id

  # AWS account IDs allowed to connect to this private link
  consumer_accounts = [
    "arn:aws:iam:123456789012:root"
  ]

  # AWS regions from which you can connect to this private link
  consumer_regions = [
    "us-west-2",
    "us-east-1"
  ]
}
