---
page_title: "hcp_consul_versions Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The Consul versions data source provides the Consul versions supported by HCP.
---

# Data Source `hcp_consul_versions`

The Consul versions data source provides the Consul versions supported by HCP.

## Example Usage

```terraform
data "hcp_consul_versions" "default" {}
```

## Schema

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **available** (List of String) The Consul versions available on HCP.
- **preview** (List of String) The preview versions of Consul available on HCP.
- **recommended** (String) The recommended Consul version for HCP clusters.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


