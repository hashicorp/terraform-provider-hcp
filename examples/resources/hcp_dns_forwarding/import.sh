# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}:{dns_forwarding_id}
terraform import hcp_dns_forwarding.example f709ec73-55d4-46d8-897d-816ebba28778:main-hvn:example-dns-forwarding
# Using the provider-default project ID, the import ID is:
# {hvn_id}:{dns_forwarding_id}
terraform import hcp_dns_forwarding.example main-hvn:example-dns-forwarding
