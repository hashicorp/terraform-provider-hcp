---
subcategory: ""
page_title: "Packer Run Tasks with Terraform - HCP Provider"
description: |-
    A guide to integrating HCP Packer with Terraform using Run Tasks. 
---

# Adding an HCP Packer Run Task to Terraform Cloud or Terraform Enterprise

You can add an HCP Packer Run Task to Terraform Cloud or Terraform Enterprise
by combining the HCP Terraform Provider and the 
[Terraform Cloud/Enterprise Provider](https://registry.terraform.io/providers/hashicorp/tfe).

## Using the `hcp_packer_run_task` Data Source

If the Terraform workspace where this config will live already has an
`hcp_packer_run_task` resource, you should use the resource instead.

```terraform
data "hcp_packer_run_task" "registry" {}

resource "tfe_organization_run_task" "hcp_packer" {
  name        = "HCP_Packer"
  description = "Ensure usage of compliant images from HCP Packer."
  enabled     = true

  url      = data.hcp_packer_run_task.registry.endpoint_url
  hmac_key = data.hcp_packer_run_task.registry.hmac_key
}
```

## Using the `hcp_packer_run_task` Resource

If the Terraform workspace where this config will live does not already have a
`hcp_packer_run_task` resource and you don't need to be able to regenerate the
HMAC Key in that workspace, you should use the data source instead.

```terraform
resource "hcp_packer_run_task" "registry" {
  regenerate_hmac = false
}

resource "tfe_organization_run_task" "hcp_packer" {
  name        = "HCP_Packer"
  description = "Ensure usage of compliant images from HCP Packer."
  enabled     = true

  url      = hcp_packer_run_task.registry.endpoint_url
  hmac_key = hcp_packer_run_task.registry.hmac_key
}
```
