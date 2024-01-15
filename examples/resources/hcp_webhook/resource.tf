resource "hcp_webhook" "example" {
  name        = "example-webhook"
  description = "My new webhook!"
  enabled     = true

  config = {
    url = "https://example.com"
  }

  subscriptions = [
    {
      events = [
        {
          actions = ["*"]
          source  = "hashicorp.packer.version"
        }
      ]
    }
  ]
}