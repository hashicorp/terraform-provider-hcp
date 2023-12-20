resource "hcp_log_streaming_destination" "example_splunk_cloud" {
  name = "example_splunk_cloud"
  splunk_cloud = {
    endpoint = "https://tenant.splunkcloud.com:8088/services/collector/event"
    token    = "someSuperSecretToken"
  }
}