# Vault Secrets AWS Integration can be imported by specifying the name of the integration
# Note that since the AWS secret access key is never returned on the Vault Secrets API,
# the next plan or apply will show a diff for that field if using the access keys authentication method.
terraform import hcp_vault_secrets_integration_aws.example my-aws-1
