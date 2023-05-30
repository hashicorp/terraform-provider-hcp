# Using an explicit project ID, the import ID is:
# {project_id}:{cluster_id}
terraform import hcp_vault_cluster.example f709ec73-55d4-46d8-897d-816ebba28778:vault-cluster
# Using the provider-default project ID, the import ID is:
# {cluster_id}
terraform import hcp_vault_cluster.example vault-cluster
