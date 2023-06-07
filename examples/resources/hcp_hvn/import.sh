# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}
terraform import hcp_hvn.example f709ec73-55d4-46d8-897d-816ebba28778:main-hvn
# Using the provider-default project ID, the import ID is:
# {hvn_id}
terraform import hcp_hvn.example main-hvn
