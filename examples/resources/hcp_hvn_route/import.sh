# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}:{hvn_route_id}
terraform import hcp_hvn_route.example f709ec73-55d4-46d8-897d-816ebba28778:main-hvn:example-hvn-route
# Using the provider-default project ID, the import ID is:
# {hvn_id}:{hvn_route_id}
terraform import hcp_hvn_route.example main-hvn:example-hvn-route
