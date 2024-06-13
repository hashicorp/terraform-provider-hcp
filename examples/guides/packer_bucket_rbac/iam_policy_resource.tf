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
