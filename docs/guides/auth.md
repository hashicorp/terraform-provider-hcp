---
subcategory: ""
page_title: "Authenticate with HCP - HCP Provider"
description: |-
    A guide to obtaining client credentials and adding them to provider configuration.
---

# Authenticate with HCP using service principal credentials

The HCP provider accepts client credentials, which are obtained on the creation of a service principal key. 

Service principals and service principal keys can be created in the HCP portal with an existing user account.

## Create a service principal

Once you have registered and logged into the HCP portal, navigate to the Access Control (IAM) page. Select the Service Principals tab and create a new service principal. Give it the role Contributor, since it will be writing resources.

## Create a service principal key

Once the service principal is created, navigate to its detail page by selecting its name in the list. Create a new service principal key.

-> **Note:** Save the client ID and secret returned on successful key creation. The client secret will not be available after creation.

## Two options to configure the provider 

Save the client ID and secret as the environment variables HCP_CLIENT_ID and HCP_CLIENT_SECRET. Or, configure the provider with the client ID and secret by copy-pasting the values directly into provider config.

```terraform
// Credentials can be set explicitly or via the environment variables HCP_CLIENT_ID and HCP_CLIENT_SECRET
provider "hcp" {
  client_id     = "service-principal-key-client-id"
  client_secret = "service-principal-key-client-secret"
}
``` 

## Optional Browser Login if no client credentials configured

The HCP provider supports log in via the browser if the client ID and client secret aren't configured. This feature is supported in provider version 0.45.0 and above. Enable browser login post update by first regenerating the provider in terminal with the following commands. 

```
cd terraform-provider-hcp
make dev
terraform init
terraform apply
```

Your web browser then will navigate to the HCP portal, where you will be prompted to login. You may now create new or manage existing resources fully authenticated.