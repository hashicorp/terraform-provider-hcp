resource "hcp_vault_secrets_integration" "example_gitlab_integration" {
  name          = "gitlab-integration"
  capabilities  = ["SYNC"]
  provider_type = "gitlab"
  gitlab_access = {
    token = "myaccesstoken"
  }
}

resource "hcp_vault_secrets_sync" "example_gitlab_project_sync" {
  name             = "gitlab-proj-sync"
  integration_name = hcp_vault_secrets_integration.example_gitlab_integration.name
  gitlab_config = {
    scope      = "PROJECT"
    project_id = "123456"
  }
}
