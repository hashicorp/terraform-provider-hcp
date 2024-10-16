resource "hcp_log_streaming_destination" "example_cloudwatch" {
  name = "example_cloudwatch"
  cloudwatch = {
    external_id    = "an-external-id"
    region         = "us-east-1"
    role_arn       = "arn:aws:iam::111111111:role/hcp-log-streaming"
    log_group_name = "a-log-group-name"
  }
}