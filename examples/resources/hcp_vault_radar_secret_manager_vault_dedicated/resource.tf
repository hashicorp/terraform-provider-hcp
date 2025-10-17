# Usage with Kubernetes auth method
resource "hcp_vault_radar_secret_manager_vault_dedicated" "secret_manager_example_1" {
  vault_url = "example-1.hashicorp.cloud:8200"
  kubernetes = {
    mount_path = "kubernetes"
    role_name  = "vault-radar-role"
  }
}

# Usage with AppRole (Push) auth method.
resource "hcp_vault_radar_secret_manager_vault_dedicated" "secret_manager_example_2" {
  vault_url = "example-2.hashicorp.cloud:8200"
  approle_push = {
    mount_path        = "approle"
    role_id_env_var   = "VAULT_APPROLE_ROLE_ID"
    secret_id_env_var = "VAULT_APPROLE_SECRET_ID"
  }
}

# Usage with Token auth method
resource "hcp_vault_radar_secret_manager_vault_dedicated" "secret_manager_example_3" {
  vault_url = "example-3.hashicorp.cloud:8200"
  token = {
    token_env_var = "VAULT_TOKEN"
  }
}