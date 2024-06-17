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
