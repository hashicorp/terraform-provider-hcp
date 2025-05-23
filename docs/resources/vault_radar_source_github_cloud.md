---
page_title: "hcp_vault_radar_source_github_cloud Resource - terraform-provider-hcp"
subcategory: "HCP Vault Radar"
description: |-
  This terraform resource manages a GitHub Cloud data source lifecycle in Vault Radar.
---

# hcp_vault_radar_source_github_cloud (Resource)

-> **Note:** This feature is currently in private beta.

This terraform resource manages a GitHub Cloud data source lifecycle in Vault Radar.

## Example Usage

```terraform
variable "github_cloud_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_source_github_cloud" "example" {
  github_organization = "my-github-org"
  token               = var.github_cloud_token
  project_id          = "my-project-id"
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `github_organization` (String) GitHub organization Vault Radar will monitor. Example: type "octocat" for the org https://github.com/octocat
- `token` (String, Sensitive) GitHub personal access token.

### Optional

- `project_id` (String) The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.

### Read-Only

- `id` (String) The ID of this resource.
