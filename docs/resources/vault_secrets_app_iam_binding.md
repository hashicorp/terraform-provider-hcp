---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: "Cloud Platform"
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

~> **Note:** `hcp_vault_secrets_app_iam_binding` can only be used if the app has an existing policy and cannot be used in conjunction with
`hcp_vault_secrets_app_iam_policy`.

## Example Usage

{{ tffile "examples/resources/hcp_vault_secrets_app_iam_binding/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}