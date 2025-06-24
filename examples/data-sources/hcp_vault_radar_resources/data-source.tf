# Returns a list of Radar resources in the project with uri matching
# that beginning with "git://github.com/hashicorp/" or "git://github.com/ibm/".
data "hcp_vault_radar_resources" "example" {
  uri_like_filter = {
    values = [
      "git://github.com/hashicorp/%",
      "git://github.com/ibm/%"
    ]
    case_insensitive = false
  }
}
