---
subcategory: ""
page_title: "Multi-Project Migration Guide - HCP Provider"
description: |-
    A guide to migrating HCP resources to multiple projects.
---

# Multi-project Migration Guide

HCP now supports multiple projects. You may migrate your HCP Terraform configuration in the following ways.

-> **Note:** Resources cannot be moved to new projects. Changing a resource's project will force its recreation. Before creating additional projects, we recommend configuring the current sole project as the provider's default project to ensure no recreation occurs.

## 1. Default to oldest project (no change required)

The HVN in this example will be created in the sole existing project, or if there is more than one project, the oldest project.

```terraform
provider "hcp" {}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}
```

## 2. Configure a default project on provider

The HVN in this example will be created in the project configured at the provider level.

```terraform
provider "hcp" {
  project_id = "f709ec73-55d4-46d8-897d-816ebba28778"
}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}
```

## 3. Configure projects on resource

The HVN will be created in its configured project, while the HCP Consul cluster will be created in its different configured project. 
Since no project is configured on the provider, the default project will be the oldest project.

```terraform
provider "hcp" {}

resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  project_id     = "f709ec73-55d4-46d8-897d-816ebba28778"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_consul_cluster" "consul_cluster" {
  cluster_id = "test-cluster"
  hvn_id     = hcp_hvn.test.hvn_id
  project_id = "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
  tier       = "development"
}
```

### Override provider project with resource project

Projects may be set at both the resource and provider level. The resource-configured project is always preferred over the provider-configured project.

```terraform
provider "hcp" {
  project_id = "f709ec73-55d4-46d8-897d-816ebba28778"
}

# This HVN will be created in the project "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  project_id     = "0f8c263e-8eb4-4a7f-a0cc-7e476afb9fd2"
  cloud_provider = "aws"
  region         = "us-west-2"
}
```

## Imports

If no project is configured at the provider level, imported resources must include their project ID to ensure uniqueness.

```shell
# terraform import {resource_type}.{resource_name} {project_id}:{hvn_id}

$ terraform import hcp_hvn.test f709ec73-55d4-46d8-897d-816ebba28778:test-hvn
```