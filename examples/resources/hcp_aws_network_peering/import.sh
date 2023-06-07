# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}:{peering_id}
terraform import hcp_aws_network_peering.peer f709ec73-55d4-46d8-897d-816ebba28778:main-hvn:11eb60b3-d4ec-5eed-aacc-0242ac120015
# Using the provider-default project ID, the import ID is:
# {hvn_id}:{peering_id}
terraform import hcp_aws_network_peering.peer main-hvn:11eb60b3-d4ec-5eed-aacc-0242ac120015
