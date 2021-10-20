---
page_title: "hcp_packer_image_iteration Data Source - terraform-provider-hcp"
subcategory: ""
description: |-
  The Packer Image data source iteration gets the most recent iteration (or build) of an image, given a channel.
---

# hcp_packer_image_iteration (Data Source)

-> **Note:** This feature is currently in beta.

The Packer Image data source iteration gets the most recent iteration (or build) of an image, given a channel.

## Example Usage

```terraform
data "hcp_packer_image_iteration" "alpine" {
  bucket_name = "alpine"
  channel     = "production"
}
```

## Schema

### Required

- **bucket_name** (String) The slug of the HCP Packer Registry image bucket to pull from.
- **channel** (String) The channel that points to the version of the image you want.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- **builds** (List of Object) Builds for this iteration. An iteration can have more than one build if it took more than one go to build all images. (see [below for nested schema](#nestedatt--builds))
- **created_at** (String) Creation time of this iteration
- **incremental_version** (Number) Incremental version of this iteration
- **organization_id** (String) The ID of the organization this HCP Packer registry is located in.
- **project_id** (String) The ID of the project this HCP Packer registry is located in.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedatt--builds"></a>
### Nested Schema for `builds`

Read-Only:

- **cloud_provider** (String) Name of the cloud provider this image is stored-in, if any.
- **component_type** (String) Name of the builder that built this. Ex: 'amazon-ebs.example'.
- **created_at** (String) Creation time of this build.
- **id** (String) HCP ID of this build.
- **images** (List of Object) (see [below for nested schema](#nestedobjatt--builds--images))
- **labels** (Map of String) Labels for this build.
- **packer_run_uuid** (String) Packer generated UUID of this build.
- **status** (String) Status of this build. DONE means that all images tied to this build were successfully built.
- **updated_at** (String) Time this build was last updated.

<a id="nestedobjatt--builds--images"></a>
### Nested Schema for `builds.images`

Read-Only:

- **created_at** (String) Creation time of this image.
- **id** (String) HCP ID of this image.
- **image_id** (String) Cloud Image ID, URL string identifying this image for the builder that built it.
- **region** (String) Region this image was built from. If any.
