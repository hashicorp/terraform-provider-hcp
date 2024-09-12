# Vault Secrets Mongo DB Atlas Integration can be imported by specifying the name of the integration
# Note that since the API private key is never returned on the Vault Secrets API,
# the next plan or apply will show a diff for that field.
terraform import hcp_vault_secrets_integration_mongodbatlas.example my-mongodbatlas-1
