# Vault Secrets Azure Integration can be imported by specifying the name of the integration
# Note that since the client secret is never returned on the Vault Secrets API,
# the next plan or apply will show a diff for that field.
terraform import hcp_vault_secrets_integration_azure.example my-azure-1
