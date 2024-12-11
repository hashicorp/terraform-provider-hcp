# Vault Secrets Integration can be imported by specifying the name of the integration
# Note that since sensitive information are never returned on the Vault Secrets API,
# the next plan or apply will show a diff for sensitive fields.
terraform import hcp_vault_secrets_integration.example my-integration-name
