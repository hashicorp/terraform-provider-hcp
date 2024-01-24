locals {
  # Map FROM bucket name TO "none"
  #
  # Empty if `var.defaultToUnassigned == false`, otherwise contains all buckets
  # not found in `var.ignoreIfNotSet` or `var.errorIfNotSet`
  defaultAssignments = var.defaultToUnassigned ? {
    for bucketName in setsubtract(keys(var.channels), setunion(var.ignoreIfNotSet, var.errorIfNotSet)) :
    bucketName => "none"
  } : {}

  # Map FROM bucket name TO null
  requireExplicitAssignments = {
    for bucketName in var.errorIfNotSet :
    bucketName => null
  }

  # Same as `var.channelLinks`, but with any buckets present in 
  # `var.explicitAssignments` removed to minimize wasted API calls
  channelLinksMinimized = {
    for bucketName, channel in var.channelLinks :
    bucketName => channel
    if !contains(keys(var.explicitAssignments), bucketName)
  }
}

data "hcp_packer_version" "channel_links" {
  for_each = local.channelLinksMinimized

  bucket_name  = each.key
  channel_name = each.value
}

locals {
  # Map FROM bucket name TO version id
  channelLinkAssignments = {
    for bucketName in keys(local.channelLinksMinimized) :
    bucketName => data.hcp_packer_version.channel_links[bucketName].fingerprint
  }

  unfilteredAssignments = merge(
    local.defaultAssignments,
    local.requireExplicitAssignments,
    local.channelLinkAssignments,
    var.explicitAssignments,
  )

  assignments = var.errorOnInvalidBucket ? local.unfilteredAssignments : {
    for bucketName, versionFingerprint in local.unfilteredAssignments :
    bucketName => versionFingerprint
    if contains(keys(var.channels), bucketName)
  }
}

resource "hcp_packer_channel_assignment" "release" {
  for_each = local.assignments

  bucket_name  = each.key
  channel_name = var.channels[each.key].name

  version_fingerprint = each.value
}