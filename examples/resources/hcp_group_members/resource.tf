resource "hcp_group_members" "example" {
  group = hcp_group.example.resource_name
  members = [
    hcp_user_principal.example1.user_id,
    hcp_user_principal.example2.user_id,
  ]
}
