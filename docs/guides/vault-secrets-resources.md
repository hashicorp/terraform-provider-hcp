---
subcategory: ""
page_title: "Manage apps and secrets with Vault Secrets"
description: |-
    Manage Vault Secrets apps and resources.
---

The HCP Provider allows you to manage your Vault Secrets apps and secrets.

The Vault Secrets app resource allows you to manage your application through the following configuration:

```terraform
resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}
```

We can also use this to create secrets based off our new application.

```terraform
resource "hcp_vault_secrets_app" "example" {
  app_name    = "example-app-name"
  description = "My new app!"
}
resource "hcp_vault_secrets_secret" "secret-example" {
  app_name     = hcp.hcp_vault_secrets_app.example.app_name
  secret_name  = "a-new-secret"
  secret_value = "a test secret"
}
```

-> **Note:** The secret value is considered sensitive and will be masked with any output. However, the secret value will be written to your state file and we recommend treating the [state file as sensitive](https://developer.hashicorp.com/terraform/language/state/sensitive-data)
