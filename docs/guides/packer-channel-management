---
subcategory: ""
page_title: "Advanced Packer Channel Management - HCP Provider"
description: |-
    A guide to integreting HCP Packer resources and data sources for more advanced channel management.
---

# Advanced Packer Channel Management

By integrating multiple HCP Packer resources and data sources, you can perform more advanced channel management tasks.

## Setting the channel assignment on a Terraform-managed channel

```terraform
resource "hcp_packer_channel" "advanced" {
  name        = "advanced"
  bucket_name = "alpine"
}

resource "hcp_packer_channel_assignment" "advanced" {
  bucket_name  = hcp_packer_channel.advanced.bucket_name
  channel_name = hcp_packer_channel.advanced.name

  # Exactly one of version, id, or fingerprint must be set:
  iteration_version = 12
  # iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
  # iteration_fingerprint = "01H1ZMW0Q2W6FT4FK27FQJCFG7"
}
```

## Setting the channel assignment to the latest complete iteration

```terraform
data "hcp_packer_iteration" "alpine_latest" {
  bucket_name = "alpine"
  channel     = "latest"
}

resource "hcp_packer_channel" "alpine_advanced" {
  name        = "advanced"
  bucket_name = data.hcp_packer_iteration.alpine_latest.bucket_name
}

resource "hcp_packer_channel_assignment" "alpine_advanced" {
  bucket_name  = hcp_packer_channel.alpine_advanced.bucket_name
  channel_name = hcp_packer_channel.alpine_advanced.name

  iteration_version = data.hcp_packer_iteration.alpine_latest.incremental_version
}
```

