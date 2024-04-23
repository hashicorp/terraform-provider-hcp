---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: "Cloud Platform"
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

!> **Be Careful!** You can accidentally lock yourself out of your vault secrets app using
this resource. Deleting a hcp_vault_secrets_app_iam_policy removes access from anyone
without organization or project level access to the app. This resource should generally only be used with apps fully managed by Terraform.
If you are trying to additively give permissions to the app, prefer using
`hcp_vault_secrets_app_iam_binding`. If you do use this resource, it is recommended to
import the policy before applying the change.

{{ .Description | trimspace }}

~> **Note:** `hcp_vault_secrets_app_iam_policy` can not be used in conjunction with
`hcp_vault_secrets_app_iam_binding`.

## Example Usage

{{ tffile "examples/resources/hcp_project_iam_policy/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the following syntax:

{{ codefile "shell" "examples/resources/hcp_vault_secrets_app_iam_policy/import.sh" }}