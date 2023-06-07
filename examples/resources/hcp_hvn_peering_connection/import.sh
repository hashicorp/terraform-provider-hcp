# Only the first HVN ID is required (hvn_1_id), HVN 2 will be populated after import.

# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_1_id}:{peering_id}
terraform import hcp_hvn_peering_connection.peer_1 f709ec73-55d4-46d8-897d-816ebba28778:hvn-1:peer-1
# Using the provider-default project ID, the import ID is:
# {hvn_1_id}:{peering_id}
terraform import hcp_hvn_peering_connection.peer_1 hvn-1:peer-1
