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

resource "hcp_packer_channel" "prod" {
  # Optionally, filter the list of bucket names before passing them to for_each
  for_each = data.hcp_packer_bucket_names.all.names

  name        = "prod"
  bucket_name = each.key
}
```

### Managing channel assignment for automatically created channels

Channel assignment for automatically created channels can be managed in many ways. Some nonexhaustive example configurations are provided below.

For examples where only one iteration identifier attribute is supported, the `iteration_id` attribute is used for example purposes. Any of the three iteration identifier attributes can be used.

You are not required to configure an assignment for every channel at the same time as automatic channel creation. Using Terraform builtin functions/expressions, it is possible to modify the examples below to select which buckets will also have channel assignments configured, even if _all_ buckets will have automatic channel creation enabled.

#### Example 1: Explicit assignments for each channel

This approach makes it easy to delegate management of each bucket's channel to a different file (or even an entirely different Terraform workspace).

##### Option 1: Individual resources

This option allows the use of different iteration identifier attributes for each channel's assignment. However, it has a lot more boilerplate than Option 2.

```terraform
resource "hcp_packer_channel_assignment" "bucket1" {
  bucket_name = "bucket1"
  # `channel_name` uses a reference to create an implicit dependency so
  # Terraform evaluates the resources in the correct order. This also
  # allows the channel name to be generated dynamically in the
  # `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod["bucket1"].name

  iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
}

resource "hcp_packer_channel_assignment" "bucket2" {
  bucket_name  = "bucket2"
  channel_name = hcp_packer_channel.prod["bucket2"].name

  iteration_fingerprint = "01H1GH3NWAK8AN98QWDSALN3VO"
}
```

##### Option 2: One resource

```terraform
resource "hcp_packer_channel_assignment" "prod" {
  for_each = {
    "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
    "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
  }

  bucket_name = each.key
  # `channel_name` uses a reference to create an implicit dependency so
  # Terraform evaluates the resources in the correct order. This also
  # allows the channel name to be generated dynamically in the
  # `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
```

#### Example 2: Assign all channels to the same value

```terraform
resource "hcp_packer_channel_assignment" "prod" {
  for_each = hcp_packer_channel.prod

  bucket_name  = each.key
  channel_name = each.value.name

  iteration_id = "01H1SF9NWAK8AP25PAWDBGZ1YD"
}
```

#### Example 3: Ensuring all channels have an assignment configuration

These approaches will ensure that all channels have an assignment configured, either with a default value or an override.

The default value used in the examples is `"none"`, which causes Terraform to set the channel to have no assigned iteration and continue to ensure that an iteration is not assigned elsewhere. Each of the three iteration identifier attributes has a "zero value" for this purpose: `iteration_id` and `iteration_fingerprint` both use `"none"`, and `iteration_version` uses `0`. 
The default value can also be set to a string literal (or integer literal, for iteration_version), a reference to an iteration identifier attribute from an `hcp_packer_iteration` data source, or any other valid iteration identifier.
If a default value is not desired, but you still want to ensure that every channel has its assignment managed by Terraform, you can replace `"none"` in the examples with `null`. This will result in a planning error if any buckets don't have an explicitly defined override.

##### Option 1a: Default with explicit overrides. Only allows overrides for valid buckets.

Overrides are provided as a map from bucket names to iteration identifiers.

This option will error if an override is set for a bucket that doesn't have a channel managed by `hcp_packer_channel.prod`. This helps to ensure that the configuration doesn't contain orphaned values.

```terraform
resource "hcp_packer_channel_assignment" "prod" {
  for_each = merge(
    { for c in hcp_packer_channel.prod : c.bucket_name => "none" },
    {
      "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
      "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
    }
  )

  bucket_name = each.key
  # Using a reference for `channel_name` allows it to be generated dynamically
  # in the `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
```

If some buckets will have their channel's assignment managed outside of Terraform (or in another Terraform workspace), you can wrap the merge function above with a filter, like this (it may be cleaner to use a local variable to store the contents of the merge function): 

```
for_each = { 
  for k, v in merge(....contents from above....) : k => v
  if !contains(["list","of","buckets","to","exclude"], k)
}
```

##### Option 1b: Default with explicit overrides. Allows overrides for invalid buckets.

Overrides are provided as a map from bucket names to iteration identifiers.

This option will not error if overrides are set for buckets that don't have a channel managed by `hcp_packer_channel.prod`. This also means it will not error if an override is configured for a bucket that does not exist in the HCP Packer registry.

```terraform
resource "hcp_packer_channel_assignment" "prod" {
  for_each = hcp_packer_channel.prod

  bucket_name  = each.key
  channel_name = each.value.name

  iteration_id = lookup(
    { # Per-bucket override for the iteration ID
      "bucket1" : "01H1SF9NWAK8AP25PAWDBGZ1YD"
      "bucket2" : "01H28NJ7WPCZA0FZZ8G3FGGTAF"
    },
    each.key,
    "none"
  )
}
```

If some buckets will have their channel's assignment managed outside of Terraform (or in another Terraform workspace), you can wrap the for_each input above with a filter, like this:

```
for_each = { 
  for k, v in hcp_packer_channel.prod : k => v
  if !contains(["list","of","buckets","to","exclude"], k)
}
```

##### Option 2: Default with filter-based overrides

Overrides are provided as a map from bucket names to iteration identifiers, generated using map comprehensions with filter functions. 
It is possible to add one or more maps containing overrides for specific buckets to the merge function in this example. Without individual overrides, this example shouldn't cause invalid bucket errors, but with individual overrides it may throw errors similar to the behavior in Option 1a.

```terraform
resource "hcp_packer_channel_assignment" "prod" {
  for_each = merge(
    { for c in hcp_packer_channel.prod : c.bucket_name => "none" },
    # If a default value is not desired, omit the line above, and buckets that 
    # aren't added to the map won't have a channel assignment set.
    # If a default value is not desired, but all channels should have an 
    # assignment set, replace `"none"` with `null` to ensure that every channel
    # is covered by at least one of the filters.
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H1SF9NWAK8AP25PAWDBGZ1YD"
      if startswith(v.bucket_name, "prefix1")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H28NK6V40TKSC4MMD3Z5NGMN"
      if startswith(v.bucket_name, "prefix2")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H28NJ7WPCZA0FZZ8G3FGGTAF"
      if endswith(v.bucket_name, "someSuffix")
    },
    {
      for c in hcp_packer_channel.prod : c.bucket_name => "01H1SF9NWAK8AP25PAWDBGZ1YD"
      if strcontains(v.bucket_name, "someContents")
    },
  )

  bucket_name = each.key
  # Using a reference for `channel_name` allows it to be generated dynamically
  # in the `hcp_packer_channel.prod` resource if desired.
  channel_name = hcp_packer_channel.prod[each.key].name

  iteration_id = each.value
}
```

If some buckets will have their channel's assignment managed outside of Terraform (or in another Terraform workspace), you can wrap the merge function above with a filter, like this (it may be cleaner to use a local variable to store the contents of the merge function): 

```
for_each = { 
  for k, v in merge(....contents from above....) : k => v
  if !contains(["list","of","buckets","to","exclude"], k)
}
```
