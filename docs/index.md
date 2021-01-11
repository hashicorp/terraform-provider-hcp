---
page_title: "hcp Provider"
subcategory: ""
description: |-
  
---

# hcp Provider



## Example Usage

```terraform
provider "hcp" {
  client_id     = "example-id"
  client_secret = "example-secret"
}
```

## Schema

### Required

- **client_id** (String) The OAuth2 Client ID for API operations.
- **client_secret** (String) The OAuth2 Client Secret for API operations.

### Optional

- **organization_id** (String) The ID of the organization for API operations.
- **project_id** (String) The ID of the project for API operations.
