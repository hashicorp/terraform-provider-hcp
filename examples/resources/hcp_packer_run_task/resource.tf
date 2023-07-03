resource "hcp_packer_run_task" "registry1" {}

# Configuring the HMAC Key to regenerate on apply
# NOTE: `regenerate_hmac` should be set to `false` (or removed from the config
# entirely) after a successful apply, to avoid constant regeneration.
resource "hcp_packer_run_task" "registry1" {
  regenerate_hmac = true
}
