resource "hcp_packer_run_task" "registry" {}

# Configuring the HMAC Key to regenerate on apply
# NOTE: While `regenerate_hmac` is set to `true` the key will be regenerated on every apply.
resource "hcp_packer_run_task" "registry" {
  regenerate_hmac = true
}
