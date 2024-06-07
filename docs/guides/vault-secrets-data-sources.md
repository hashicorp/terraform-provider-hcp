---
subcategory: ""
page_title: "Using the Vault Secrets Datasources"
description: |-
    Fetch your secrets from Vault Secrets
---

The HCP Provider offers two ways of accessing secrets as a data source. The first is through fetching an app’s worth of secrets and the second is through fetching a singular secret by name.

Below is an example of fetching an app’s worth of secrets and accessing this throughout your Terraform configuration.

```terraform
data "hcp_vault_secrets_app" "app-data" {
  app_name = "example-vault-secrets-app"
}

resource "example_resource" "example" {
  example_attr = data.hcp_vault_secrets_secret.secrets["my_secret_key"]
}
```

We also allow you to fetch a singular secret from Vault Secrets.

```terraform
data "hcp_vault_secrets_secret" "secret-data" {
  app_name    = "example-vault-secrets-app"
  secret_name = "my_secret"
}

resource "example_resource" "example" {
  example_attr = data.hcp_vault_secrets_secret.secret_value
}
```

If the secret you are trying to fetch is known to be of type 'rotating', then using the following data source will give a more granual access to the secret's nested data:

```terraform
data "hcp_vault_secrets_rotating_secret" "secret-data" {
  app_name    = "example-vault-secrets-app"
  secret_name = "my_secret"
}

resource "example_resource" "example" {
  example_attr1 = data.hcp_vault_secrets_rotating_secret.secret_values["username"]
  example_attr2 = data.hcp_vault_secrets_rotating_secret.secret_values["password"]
}
```
