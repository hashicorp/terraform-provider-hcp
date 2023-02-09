resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration {
    id = "iteration-id"
  }
}

# Update assigned iteration using an iteration fingerprint
resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration {
    fingerprint = "fingerprint-associated-to-iteration"
  }
}

# Update assigned iteration using an iteration incremental version
resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration {
    // incremental_version is the version number assigned to a completed iteration.
    incremental_version = 1
  }
}

