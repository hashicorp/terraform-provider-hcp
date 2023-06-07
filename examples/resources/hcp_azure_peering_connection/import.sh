# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}:{peering_id}
terraform import hcp_azure_peering_connection.peer f709ec73-55d4-46d8-897d-816ebba28778:main-hvn:199e7e96-4d5f-4456-91f3-b6cc71f1e561
# Using the provider-default project ID, the import ID is:
# {hvn_id}:{peering_id}
terraform import hcp_azure_peering_connection.peer main-hvn:199e7e96-4d5f-4456-91f3-b6cc71f1e561
