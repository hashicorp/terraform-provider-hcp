# References

This documentation also contains reference material specific to certain functionality.

_Note: This documentation is intended for Terraform HCP Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

## Beta Features

The HCP team will sometimes use a pattern where a feature or set of features are only available to a limited group of beta customers. This allows the team to develop and test these features against a smaller group before releasing them more broadly. Features that are in a beta phase are generally controlled at the service level through feature flags.

In order to support Terraform usage for our beta customers, **the current pattern is to make the Terraform resources publicly available, and use a banner in the resource documentation to indicate that the feature is only available to beta customers.** Other users that attempt to use the feature with Terraform will be presented with an error message from the service indicating that the feature is not enabled.

Here are some steps that can be followed when adding resources that are only available to beta customers:
1. Ensure there is a feature flag at the service level that will produce a clear error message for a Terraform user without access
1. Create a template file for the resource documentation to include a banner. For example, `templates/resources/aws_transit_gateway_attachment.md.tmpl` used this content:
```
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} `{{.Type}}`

-> **Note:** This feature is currently in private beta. If you would like early access, please [contact our sales team](https://www.hashicorp.com/contact-sales).

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/resources/hcp_aws_transit_gateway_attachment/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

-> **Note:** When importing a transit gateway attachment, you will want to configure a `lifecycle` configuration block with an `ignore_changes` argument including `resource_share_arn`. This is needed because its value is no longer retrievable after creation.

Import is supported using the following syntax:

{{ codefile "shell" "examples/resources/hcp_aws_transit_gateway_attachment/import.sh" }}

```
1. Merge the resources into `main` like normal, and include them in the next public release of the provider
1. Once the feature is no longer in a beta phase, remove the banner by deleting the custom template file
