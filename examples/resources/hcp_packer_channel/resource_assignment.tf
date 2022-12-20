resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration_assignment {
    id = "iteration-id"
  }
}

# Update assigned iteration using an iteration fingerprint
resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration_assignment {
    fingerprint = "fingerprint-associated-to-iteration"
  }
}

# Update assigned iteration using an iteration incremental version
resource "hcp_packer_channel" "staging" {
  name        = "staging"
  bucket_name = "alpine"
  iteration_assignment {
    // incremental_version is the version number assigned to a completed iteration.
    incremental_version = 1
  }
}

