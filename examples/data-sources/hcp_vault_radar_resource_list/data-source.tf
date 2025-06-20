# Returns a list of Radar resources in the project with uri matching
# that beginning with "git://github.com/hashicorp/" or "git://github.com/ibm/".
data "hcp_vault_radar_resource_list" "example" {
  uri_like_filter = [
    "git://github.com/hashicorp/%",
    "git://github.com/ibm/%"
  ]
}
