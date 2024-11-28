# Vault Secrets GCP Integration can be imported by specifying the name of the integration
# Note that since the service account credentials are never returned on the Vault Secrets API,
# the next plan or apply will show a diff for that field if using the service account key authentication method.
terraform import hcp_vault_secrets_integration_gcp.example my-gcp-1
