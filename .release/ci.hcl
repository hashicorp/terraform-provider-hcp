# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

schema = "2"

project "terraform-provider-hcp" {
  team = "hcp"

  slack {
    notification_channel = "C05BBTQFCGZ"
  }

  github {
    organization     = "hashicorp"
    repository       = "terraform-provider-hcp"
    release_branches = ["main"]
  }
}

event "merge" {
}

event "build" {
  action "build" {
    depends = ["merge"]

    organization = "hashicorp"
    repository   = "terraform-provider-hcp"
    workflow     = "build"
  }
}

event "prepare" {
  depends = ["build"]

  action "prepare" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "prepare"
    depends      = ["build"]
  }

  notification {
    on = "fail"
  }
}

event "trigger-staging" {
}

event "promote-staging" {
  action "promote-staging" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-staging"
    depends      = null
    config       = "release-metadata.hcl"
  }

  depends = ["trigger-staging"]

  notification {
    on = "always"
  }
}

event "trigger-production" {
}

event "promote-production" {
  action "promote-production" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-production"
    depends      = null
    config       = ""
  }

  depends = ["trigger-production"]

  notification {
    on = "always"
  }
}

event "bump-version-patch" {
  depends = ["promote-production-packaging"]
  action "bump-version" {
    organization = "hashicorp"
    repository = "crt-workflows-common"
    workflow = "bump-version"
  }
  notification {
    on = "fail"
  }
}
