---
page_title: "hcp_vault_radar_integration_jira_connection Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  This terraform resource manages an Integration Jira Connection in Vault Radar.
---

# hcp_vault_radar_integration_jira_connection (Resource)

-> **Note:** HCP Vault Radar Terraform resources are in preview.

This terraform resource manages an Integration Jira Connection in Vault Radar.

## Example Usage

```terraform
variable "jira_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_integration_jira_connection" "jira_connection" {
  name     = "example connection to jira"
  email    = "jane.smith@example.com"
  token    = var.jira_token
  base_url = "https://example.atlassian.net"
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `base_url` (String) The Jira base URL. Example: https://acme.atlassian.net
- `email` (String, Sensitive) Jira user's email.
- `name` (String) Name of connection. Name must be unique.
- `token` (String, Sensitive) A Jira API token.

### Optional

- `project_id` (String) The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.

### Read-Only

- `id` (String) The ID of this resource.
