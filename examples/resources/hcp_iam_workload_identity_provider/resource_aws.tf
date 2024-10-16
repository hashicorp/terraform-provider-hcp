resource "hcp_service_principal" "workload_sp" {
  name = "my-app-runtime"
}

resource "hcp_iam_workload_identity_provider" "example" {
  name              = "aws-example"
  service_principal = hcp_service_principal.workload_sp.resource_name
  description       = "Allow my-app on AWS to act as my-app-runtime service principal"

  aws {
    # Only allow workloads from this AWS Account to exchange identity
    account_id = "123456789012"
  }

  # Only allow workload's running with the correct AWS IAM Role
  conditional_access = "aws.arn matches `^arn:aws:sts::123456789012:assumed-role/my-app-role`"
}
