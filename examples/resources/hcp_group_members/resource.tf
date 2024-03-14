resource "hcp_group_members" "example" {
  group = hcp_group.example.id
  members = [
    hcp_user_principal.example1.id,
    hcp_user_principal.example2.id,
  ]
}
