resource "hcp_notifications_webhook" "example" {
  name        = "example-webhook"
  description = "Notify for all of the events for all Packer artifact versions existing in the project."

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