---
page_title: "Resource hcp_vault_secrets_rotating_secret - terraform-provider-hcp"
subcategory: "HCP Vault Secrets"
description: |-
  The Vault Secrets rotating secret resource manages a rotating secret configuration.
---

# hcp_vault_secrets_rotating_secret (Resource)

The Vault Secrets rotating secret resource manages a rotating secret configuration.

## Example Usage

```terraform
resource "hcp_vault_secrets_rotating_secret" "example_aws" {
  app_name             = "my-app-1"
  secret_provider      = "aws"
  name                 = "my_aws_1"
  integration_name     = "my-aws-1"
  rotation_policy_name = "built-in:60-days-2-active"
  aws_access_keys = {
    iam_username = "my-iam-username"
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_gcp" {
  app_name             = "my-app-1"
  secret_provider      = "gcp"
  name                 = "my_gcp_1"
  integration_name     = "my-gcp-1"
  rotation_policy_name = "built-in:60-days-2-active"
  gcp_service_account_key = {
    service_account_email = "<name>>@<project>.iam.gserviceaccount.com"
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_mongodb_atlas" {
  app_name             = "my-app-1"
  secret_provider      = "mongodb_atlas"
  name                 = "my_mongodb_atlas_1"
  integration_name     = "my-mongodbatlas-1"
  rotation_policy_name = "built-in:60-days-2-active"
  mongodb_atlas_user = {
    project_id    = "<uuid>>"
    database_name = "my-cluster-0"
    roles         = ["readWrite", "read"]
  }
}

resource "hcp_vault_secrets_rotating_secret" "example_twilio" {
  app_name             = "my-app-1"
  secret_provider      = "twilio"
  name                 = "my_twilio_1"
  integration_name     = "my-twilio-1"
  rotation_policy_name = "built-in:60-days-2-active"
  twilio_api_key       = {}
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_name` (String) Vault Secrets application name that owns the secret.
- `integration_name` (String) The Vault Secrets integration name with the capability to manage the secret's lifecycle.
- `name` (String) The Vault Secrets secret name.
- `rotation_policy_name` (String) Name of the rotation policy that governs the rotation of the secret.
- `secret_provider` (String) The third party platform the dynamic credentials give access to. One of `aws` or `gcp`.

### Optional

- `aws_access_keys` (Attributes) AWS configuration to manage the access key rotation for the given IAM user. Required if `secret_provider` is `aws`. (see [below for nested schema](#nestedatt--aws_access_keys))
- `gcp_service_account_key` (Attributes) GCP configuration to manage the service account key rotation for the given service account. Required if `secret_provider` is `gcp`. (see [below for nested schema](#nestedatt--gcp_service_account_key))
- `mongodb_atlas_user` (Attributes) MongoDB Atlas configuration to manage the user password rotation on the given database. Required if `secret_provider` is `mongodb_atlas`. (see [below for nested schema](#nestedatt--mongodb_atlas_user))
- `project_id` (String) HCP project ID that owns the HCP Vault Secrets integration. Inferred from the provider configuration if omitted.
- `twilio_api_key` (Attributes) Twilio configuration to manage the api key rotation on the given account. Required if `secret_provider` is `twilio`. (see [below for nested schema](#nestedatt--twilio_api_key))

### Read-Only

- `organization_id` (String) HCP organization ID that owns the HCP Vault Secrets integration.

<a id="nestedatt--aws_access_keys"></a>
### Nested Schema for `aws_access_keys`

Required:

- `iam_username` (String) AWS IAM username to rotate the access keys for.


<a id="nestedatt--gcp_service_account_key"></a>
### Nested Schema for `gcp_service_account_key`

Required:

- `service_account_email` (String) GCP service account email to impersonate.


<a id="nestedatt--mongodb_atlas_user"></a>
### Nested Schema for `mongodb_atlas_user`

Required:

- `database_name` (String) MongoDB Atlas database or cluster name to rotate the username and password for.
- `project_id` (String) MongoDB Atlas project ID to rotate the username and password for.
- `roles` (List of String) MongoDB Atlas roles to assign to the rotating user.


<a id="nestedatt--twilio_api_key"></a>
### Nested Schema for `twilio_api_key`