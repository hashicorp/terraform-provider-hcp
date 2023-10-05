---
subcategory: ""
page_title: "Advanced Packer Channel Management - HCP Provider"
description: |-
    A guide to integreting HCP Packer resources and data sources for more advanced channel management.
---

# Advanced Packer Channel Management

You can integrate multiple HCP Packer resources and data sources to perform advanced channel management tasks.

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

## Automatically creating channels for new and existing buckets

```terraform
data "hcp_packer_bucket_names" "all" {}

resource "hcp_packer_channel" "release" {
  for_each = data.hcp_packer_bucket_names.all.names

  name        = "release"
  bucket_name = each.key
}
```

Optionally, you can use Terraform functions/expressions to filter the list of bucket names before providing it to `for_each` if channels should only be created for a subset of buckets.

### Managing channel assignment for automatically created channels

Channel assignment for automatically created channels can be managed in many ways. A nonexhaustive example configuration is provided below. 

The `iteration_id` attribute is used for example purposes, but any of the three iteration identifier attributes can be used.

```terraform
data "hcp_packer_iteration" "all_latest" {
  for_each = data.hcp_packer_bucket_names.all.names

  bucket_name = each.key
  channel     = "latest"
}

data "hcp_packer_iteration" "bucket3_staging" {
  bucket_name = "bucket3"
  channel     = "staging"
}

resource "hcp_packer_channel_assignment" "release" {
  for_each = merge(
    { # Defaulting all channels to be unassigned and Terraform-managed
      for bucketName in keys(hcp_packer_channel.release) :
      bucketName => "none"
    },
    { # Assigning channels that match a filter to an iteration fetched from another channel
      for bucketName in keys(hcp_packer_channel.release) :
      bucketName => data.hcp_packer_iteration.all_latest[bucketName].id
      if endswith(bucketName, "-dev")
    },
    { # Individual channel assignments
      "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
      "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
      "bucket3" : data.hcp_packer_iteration.bucket3_staging.id
    }
  )

  bucket_name = each.key
  # Using a reference for `channel_name` allows it to be generated dynamically
  # in the `hcp_packer_channel.release` resource if desired.
  channel_name = hcp_packer_channel.release[each.key].name

  iteration_id = each.value
}
```

You are not required to configure an assignment for every channel at the same time as automatic channel creation. Using Terraform builtin functions/expressions, it is possible to limit which channels should have assignments configured by default.

The default value used in the example is `"none"`, which causes Terraform to set the channel to have no assigned iteration and continue to ensure that an iteration is not assigned elsewhere. Each of the three iteration identifier attributes has a "zero value" for this purpose: `iteration_id` and `iteration_fingerprint` both use `"none"`, and `iteration_version` uses `0`. The default value can also be set to `null` if you want to ensure that every channel has an assignment defined explicitly rather than using a default.

If an invalid bucket name is provided in the `for_each` map, an error will be thrown. This helps to ensure that the configuration doesn't contain orphaned values, but can cause plan failures when buckets are deleted. If this behavior is undesirable, filter out invalid buckets from the result of the `merge` function.

An [example module](https://github.com/hashicorp/terraform-provider-hcp/tree/main/examples/guides/packer_channel_management/bucket_names/example_module) is available that includes options to leave select channels unmanaged, require explicit configurations for select channels, ignore invalid bucket names in the configuration, and automatically assign an iteration fetched from another channel.