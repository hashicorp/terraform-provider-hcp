---
subcategory: ""
page_title: "Managing HCP Packer Bucket IAM Policies"
description: |-
    A guide to using HCP Packer bucket resource along with binding or policy resource to manage bucket level access.
---

# Managing HCP Packer Bucket IAM Policies

You can grant specific users, service principals, or groups contributor or admin level access to a specific HCP Packer bucket using either a `hcp_packer_bucket_iam_binding` or `hcp_packer_bucket_iam_policy` resource.  Whenever a user is invited to a project they will have read level access to all resources, but you can restrict which of the principals in your project can maintain specific buckets.

A resource's policy is a list of bindings to assign roles to multiple users, groups, or service principals. The `hcp_packer_bucket_iam_policy` resource sets the Bucket IAM policy and replaces any existing policy.

The following example assigns the role `contributor` to a user principal and a service principal for the `production` bucket.

```terraform
data "hcp_iam_policy" "mypolicy" {
  bindings = [
    {
      role = "roles/contributor"
      principals = [
        "user-principal-id-1",
        "service-principal-id-1",
      ]
    },
  ]
}

resource "hcp_packer_bucket" "production" {
  name = "production"
}

resource "hcp_packer_bucket_iam_policy" "example" {
  resource_name = hcp_packer_bucket.production.resource_name
  policy_data   = data.hcp_iam_policy.mypolicy.policy_data
}
```

The following example assigns role contriubtor for a service principal to the production bucket, and also preserves existing bindings.

```terraform
resource "hcp_service_principal" "my-sp" {
  name = "my-sp"
}

resource "hcp_packer_bucket" "production" {
  name = "production"
}

resource "hcp_packer_bucket_iam_binding" "example" {
  resource_name = hcp_packer_bucket.production.resource_name
  principal_id  = hcp_service_principal.my-sp.resource_id
  role          = "roles/contributor"
}
```
