resource "hcp_log_streaming_destination" "example_splunk_cloud" {
  name = "example_splunk_cloud"
  splunk_cloud = {
    endpoint = "https://http-inputs-tenant.splunkcloud.com:443/services/collector/event"
    token    = "someSuperSecretToken"
  }
}