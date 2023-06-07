# Using an explicit project ID, the import ID is:
# {project_id}:{bucket_name}:{channel_name}
terraform import hcp_packer_channel.staging f709ec73-55d4-46d8-897d-816ebba28778:alpine:staging
# Using the provider-default project ID, the import ID is:
# {bucket_name}:{channel_name}
terraform import hcp_packer_channel.staging alpine:staging
