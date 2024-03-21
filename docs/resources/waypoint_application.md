---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hcp_waypoint_application Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  Waypoint Application resource
---

# hcp_waypoint_application (Resource)

Waypoint Application resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `application_template_id` (String) ID of the Application Template this Application is based on.
- `name` (String) The name of the Application.

### Optional

- `project_id` (String) The ID of the HCP project where the Waypoint Application is located.
- `readme_markdown` (String) Instructions for using the Application (markdownformat supported). Note: this is a base64 encoded string, andcan only be set in configuration after initial creation. Theinitial version of the README is generated from the READMETemplate from source Application Template.

### Read-Only

- `application_template_name` (String) Name of the Application Template this Application is based on.
- `id` (String) The ID of the Application.
- `namespace_id` (String) Internal Namespace ID.
- `organization_id` (String) The ID of the HCP organization where the Waypoint Application is located.