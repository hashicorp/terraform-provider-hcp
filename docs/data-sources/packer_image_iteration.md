---
page_title: "hcp_packer_image_iteration Data Source - terraform-provider-hcp"
subcategory: "HCP Packer"
description: |-
  The Packer ImageIteration data source iteration gets the most recent iteration (or build) of an image, given a channel.
---

# hcp_packer_image_iteration (Data Source)

-> **Note:** The `hcp_packer_image_iteration` data source is deprecated. Use the `hcp_packer_image` or `hcp_packer_iteration` data sources instead.

The Packer ImageIteration data source iteration gets the most recent iteration (or build) of an image, given a channel.

## Example Usage

```terraform
data "hcp_packer_image_iteration" "alpine" {
  bucket_name = "alpine"
  channel     = "production"
}
```

## Schema

### Required

- `bucket_name` (String) The slug of the HCP Packer Registry bucket to pull from.
- `channel` (String) The channel that points to the version of the image you want.

### Optional

- `project_id` (String) The ID of the HCP project where the HCP Packer registry is located. If not specified, the project specified in the HCP Provider config block will be used, if configured. If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `builds` (List of Object) Builds for this iteration. An iteration can have more than one build if it took more than one go to build all images. (see [below for nested schema](#nestedatt--builds))
- `created_at` (String) Creation time of this iteration
- `id` (String) The ID of this resource.
- `incremental_version` (Number) Incremental version of this iteration
- `organization_id` (String) The ID of the organization this HCP Packer registry is located in.
- `revoke_at` (String) The revocation time of this iteration. This field will be null for any iteration that has not been revoked or scheduled for revocation.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `default` (String)


<a id="nestedatt--builds"></a>
### Nested Schema for `builds`

Read-Only:

- `cloud_provider` (String) Name of the cloud provider this image is stored-in, if any.
- `component_type` (String) Name of the builder that built this. Ex: 'amazon-ebs.example'.
- `created_at` (String) Creation time of this build.
- `id` (String) HCP ID of this build.
- `images` (List of Object) (see [below for nested schema](#nestedobjatt--builds--images))
- `labels` (Map of String) Labels for this build.
- `packer_run_uuid` (String) Packer generated UUID of this build.
- `status` (String) Status of this build. DONE means that all images tied to this build were successfully built.
- `updated_at` (String) Time this build was last updated.

<a id="nestedobjatt--builds--images"></a>
### Nested Schema for `builds.images`

Read-Only:

- `created_at` (String) Creation time of this image.
- `id` (String) HCP ID of this image.
- `image_id` (String) Cloud Image ID, URL string identifying this image for the builder that built it.
- `region` (String) Region this image was built from. If any.
