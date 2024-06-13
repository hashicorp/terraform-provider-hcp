---
subcategory: ""
page_title: "Managing Packer Bucket IAM Policies"
description: |-
    A guide to using HCP Packer bucket resource along with binding or policy resource to manage bucket level access.
---

# Managing Packer Bucket IAM Policies

Using the `hcp_packer_bucket` resource along side either the `hcp_packer_bucket_iam_policy` resource with policy data from an `hcp_iam_policy` data source or `hcp_packer_bucket_iam_binding` resource along side a service principal ID.

A resource's policy is a list of bindings to assign roles to multiple users, groups, or service principals. The `hcp_packer_bucket_iam_policy` resource sets the Bucket IAM policy and replaces any existing policy.

The following example assigns the role `contributor` to a user principal and a service principal for the `production` bucket.

```
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

A resource's policy binding assigns a single role to a list of principals. The `hcp_packer_bucket_iam_binding` resource updates the Bucket IAM policy to bind a role to a new member. Existing bindings are preserved.

The following example assigns role contriubtor for a service principal to the production bucket, and also preserves existing bindings.

```
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