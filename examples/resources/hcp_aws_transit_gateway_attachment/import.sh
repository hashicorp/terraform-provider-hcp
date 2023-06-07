# Using an explicit project ID, the import ID is:
# {project_id}:{hvn_id}:{transit_gateway_attachment_id}
terraform import hcp_aws_transit_gateway_attachment.example f709ec73-55d4-46d8-897d-816ebba28778:main-hvn:example-tgw-attachment
# Using the provider-default project ID, the import ID is:
# {hvn_id}:{transit_gateway_attachment_id}
terraform import hcp_aws_transit_gateway_attachment.example main-hvn:example-tgw-attachment
