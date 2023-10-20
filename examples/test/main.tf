terraform {
  required_providers {
    hcp = {
      source  = "localhost/providers/hcp"
      version = "0.0.1"
    }
  }
}

provider "hcp" {
  project_id="0b880e01-c47f-4d95-b3e5-c5f2afe5bff0"
}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_consul_cluster" "example" {
 cluster_id = "consul-cluster-test"
 hvn_id     = hcp_hvn.test.hvn_id
 tier       = "development"
 lifecycle {
   prevent_destroy = true
 }
}

resource "hcp_consul_cluster_root_token" "example" {
  cluster_id = "consul-cluster"
}

resource "hcp_consul_snapshot" "example" {
  cluster_id    = "consul-cluster"
  snapshot_name = "my-snapshot"
}

data "hcp_consul_agent_kubernetes_secret" "test" {
  cluster_id = hcp_consul_cluster.example.cluster_id
}

data "hcp_consul_cluster" "example" {
  cluster_id = hcp_consul_cluster.example.cluster_id
}

data "hcp_consul_versions" "default" {}

resource "hcp_vault_cluster" "example" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.test.hvn_id
  tier       = "dev"
}

data "hcp_vault_cluster" "example" {
  cluster_id = hcp_vault_cluster.example.cluster_id
}

resource "hcp_vault_cluster_admin_token" "example" {
  cluster_id = hcp_vault_cluster.example.cluster_id
}

# resource "hcp_packer_channel" "advanced" {
#   name        = "advanced"
#   bucket_name = "alpine"
# }

resource "hcp_boundary_cluster" "example" {
  cluster_id = "boundary-cluster"
  username   = "test-user"
  password   = "Password123!"
  tier = "standard"
}

data "hcp_boundary_cluster" "example" {
  cluster_id = hcp_boundary_cluster.example.cluster_id
}

resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}

resource "hcp_vault_secrets_secret" "example" {
  app_name     = "example-app-name"
  secret_name  = "example_secret"
  secret_value = "hashi123"
}

data "hcp_vault_secrets_app" "example" {
  app_name = "example-app-name"
}

data "hcp_vault_secrets_secret" "example" {
  app_name    = "example-app-name"
  secret_name = "example_secret"
}
