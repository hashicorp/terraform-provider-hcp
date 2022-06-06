resource "hcp_hvn" "example" {
  hvn_id         = "hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "hcp_vault_cluster" "example" {
  cluster_id = "vault-cluster"
  hvn_id     = hcp_hvn.example.hvn_id
  tier       = "standard_large"
  metrics_config {
    datadog_api_key = "test_datadog"
    datadog_region =       "us1"
  }
  audit_log_config {
    datadog_api_key = "test_datadog"
    datadog_region  = "us1"
  }
  lifecycle {
    prevent_destroy = true
  }
}
