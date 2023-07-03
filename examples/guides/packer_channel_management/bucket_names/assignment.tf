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
