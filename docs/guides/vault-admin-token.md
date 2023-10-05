---
subcategory: ""
page_title: "Create a Vault cluster and admin token - HCP Provider"
description: |-
    An example of creating a Vault cluster and admin token.
---

# Create a new Vault cluster and an admin token

Once you have an HVN, HCP Vault enables you to quickly deploy a Vault Enterprise cluster in AWS across a variety of environments while offloading the operations burden to the SRE experts at HashiCorp.
The cluster's admin token grants its bearer administrator access to the Vault cluster. This admin token is valid for six hours. On subsequent reads after creation, 
the resource will check if the admin token is close to expiration or expired and automatically refresh as needed.

```terraform
resource "hcp_vault_cluster" "example_vault_cluster" {
  hvn_id     = hcp_hvn.example_hvn.hvn_id
  cluster_id = "hcp-tf-example-vault-cluster"
}

resource "hcp_vault_cluster_admin_token" "example_vault_admin_token" {
  cluster_id = hcp_vault_cluster.example_vault_cluster.cluster_id
}
```
