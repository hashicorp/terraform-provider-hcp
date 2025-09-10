resource "hcp_service_principal" "example" {
  name = "example-sp"
}

# Note this requires the Terraform to be run regularly
resource "time_rotating" "key_rotation" {
  rotation_days = 14
}

resource "hcp_service_principal_key" "key" {
  service_principal = hcp_service_principal.example.resource_name
  rotate_triggers = {
    rotation_time = time_rotating.key_rotation.rotation_rfc3339
  }
}
