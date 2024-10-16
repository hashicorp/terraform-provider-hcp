resource "hcp_log_streaming_destination" "example_datadog" {
  name = "example_datadog"
  datadog = {
    endpoint        = "https://datadog-api.com"
    api_key         = "API_KEY_VALUE_HERE"
    application_key = "APPLICATION_VALUE_HERE"
  }
}