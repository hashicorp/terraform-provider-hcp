package webhook_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	webhookmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-webhook/preview/2023-05-31/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccNotificationsWebhookResource(t *testing.T) {
	projectID := os.Getenv("HCP_PROJECT_ID")
	webhookName := acctest.RandString(16)
	updatedWebhookName := acctest.RandString(16)
	webhookURL := "https://" + acctest.RandString(10) + ".com"
	webhookDescription := acctest.RandString(200)

	hmac := acctest.RandString(16)
	updatedHmac := acctest.RandString(16)

	fmt.Println("webhook name " + webhookName)
	fmt.Println("webhook update name " + updatedWebhookName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test that basic disabled webhook can be created with the minimum required fields
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
			},
			{
				// Test that webhook can be imported into state
				ResourceName:                         "hcp_notifications_webhook.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccWebhookImportID,
				ImportStateVerify:                    true,
			},
			{
				// Test that webhook can be updated
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					WithDescription(webhookDescription).
					WithSubscriptions([]webhookmodels.HashicorpCloudWebhookWebhookSubscription{
						{
							ResourceID: "some_resource_id",
							Events: []*webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent{
								{
									Action: "update",
									Source: "hashicorp.packer.version",
								},
							},
						},
					}).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "description", webhookDescription),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.0.resource_id", "some_resource_id"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.0.events.0.actions.0", "update"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.0.events.0.source", "hashicorp.packer.version"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
			},
			{
				// Test that webhook hmac key can be updated
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					WithHmac(hmac).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.hmac_key", hmac),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
			},
			{
				// Test that webhook hmac key can be updated to another hmac key
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					WithHmac(updatedHmac).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.hmac_key", updatedHmac),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
			},
			{
				// Test that webhook name can be updated
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(updatedWebhookName).
					WithURL(webhookURL).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", updatedWebhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.#", "0"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, updatedWebhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
			},
			{
				// Test that webhook name is updated even when other updates fail
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					WithSubscriptions([]webhookmodels.HashicorpCloudWebhookWebhookSubscription{
						{
							ResourceID: "some_resource_id",
							Events: []*webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent{
								{
									Action: "revoke",
									Source: "hashicorp.packer.version",
								},
							},
						},
						{
							ResourceID: "some_resource_id", // same resource id should fail to update
							Events: []*webhookmodels.HashicorpCloudWebhookWebhookSubscriptionEvent{
								{
									Action: "revoke",
									Source: "hashicorp.packer.version",
								},
							},
						},
					}).
					// If enabled it will fail to create the webhook since we don't have a valid url to provide
					WithEnabled(false).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.#", "0"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
				ExpectError: regexp.MustCompile(`.*duplicated resource subscription found.*`),
			},
			{
				// Test that trying to enable webhook with invalid url fails creation
				Config: NewWebhookResourceConfigBuilder("example").
					WithName(webhookName).
					WithURL(webhookURL).
					WithEnabled(true).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "name", webhookName),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "config.url", webhookURL),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "enabled", "false"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "subscriptions.#", "0"),
					resource.TestCheckResourceAttr("hcp_notifications_webhook.example", "resource_name",
						fmt.Sprintf("webhook/project/%s/geo/us/webhook/%s", projectID, webhookName)),
					resource.TestCheckResourceAttrSet("hcp_notifications_webhook.example", "resource_id"),
				),
				ExpectError: regexp.MustCompile(`.*Error verifying webhook configuration.*`),
			},
		},
	})
}

// testAccWebhookImportID retrieves the resource_name so that it can be imported.
func testAccWebhookImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_notifications_webhook.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_name"]
	if !ok {
		return "", fmt.Errorf("resource_name not set")
	}

	return id, nil
}

type WebhookResourceConfigBuilder struct {
	resourceName  string
	name          string
	url           string
	hmac          string
	projectID     string
	description   string
	enabled       *bool
	subscriptions []webhookResourceConfigSubscription
}

type webhookResourceConfigSubscription struct {
	resourceID string
	events     []webhookResourceConfigSubscriptionEvent
}

type webhookResourceConfigSubscriptionEvent struct {
	action string
	source string
}

func NewWebhookResourceConfigBuilder(name string) WebhookResourceConfigBuilder {
	return WebhookResourceConfigBuilder{
		resourceName: name,
	}
}

func (b WebhookResourceConfigBuilder) WithName(name string) WebhookResourceConfigBuilder {
	b.name = name
	return b
}

func (b WebhookResourceConfigBuilder) WithURL(url string) WebhookResourceConfigBuilder {
	b.url = url
	return b
}

func (b WebhookResourceConfigBuilder) WithHmac(hmac string) WebhookResourceConfigBuilder {
	b.hmac = hmac
	return b
}

func (b WebhookResourceConfigBuilder) WithProjectID(projectID string) WebhookResourceConfigBuilder {
	b.projectID = projectID
	return b
}

func (b WebhookResourceConfigBuilder) WithDescription(description string) WebhookResourceConfigBuilder {
	b.description = description
	return b
}

func (b WebhookResourceConfigBuilder) WithEnabled(enabled bool) WebhookResourceConfigBuilder {
	b.enabled = &enabled
	return b
}

// WithSubscriptions only accept subscriptions with one event per source to simplify test code logic.
func (b WebhookResourceConfigBuilder) WithSubscriptions(subscriptions []webhookmodels.HashicorpCloudWebhookWebhookSubscription) WebhookResourceConfigBuilder {
	subs := make([]webhookResourceConfigSubscription, len(subscriptions))

	for i, subscription := range subscriptions {
		sub := webhookResourceConfigSubscription{}
		if subscription.ResourceID != "" {
			sub.resourceID = subscription.ResourceID
		}

		events := make([]webhookResourceConfigSubscriptionEvent, len(subscription.Events))
		for j, event := range subscription.Events {
			events[j] = webhookResourceConfigSubscriptionEvent{
				action: event.Action,
				source: event.Source,
			}
		}

		if len(events) > 0 {
			sub.events = events
		}

		subs[i] = sub
	}

	b.subscriptions = subs
	return b
}

func (b WebhookResourceConfigBuilder) Build() string {
	var enabled, subscriptions string

	if b.enabled != nil {
		enabled = fmt.Sprintf(`enabled = %t`, *b.enabled)
	}

	if len(b.subscriptions) > 0 {
		subscriptions = "subscriptions = ["

		for _, sub := range b.subscriptions {
			var events string

			if len(sub.events) > 0 {
				events = "events = ["

				for _, e := range sub.events {
					event := fmt.Sprintf(`
				{
					actions = [%q]
					source = %q
				},`, e.action, e.source)
					events = fmt.Sprintf(`%s %s`, events, event)
				}
				// remove last comma

				events = events[:len(events)-1]
				events = fmt.Sprintf(`%s ]`, events)
			}

			subscription := fmt.Sprintf(`
			{
				resource_id = %q
				%s
			},
			`, sub.resourceID, events)
			subscriptions = fmt.Sprintf(`%s %s`, subscriptions, subscription)
		}

		// remove last comma
		subscriptions = subscriptions[:len(subscriptions)-1]
		subscriptions = fmt.Sprintf(`%s ]`, subscriptions)
	}

	webhookConfig := fmt.Sprintf(`
	config = {
		url = %q
`, b.url)
	if b.hmac != "" {
		webhookConfig = fmt.Sprintf(`
		%s
		hmac_key = %q
`, webhookConfig, b.hmac)
	}
	webhookConfig = fmt.Sprintf("%s }", webhookConfig)

	config := fmt.Sprintf(`
resource "hcp_notifications_webhook" "%s" {
	name = %q
	description = %q
	
	%s
	%s
	%s
}`,
		b.resourceName,
		b.name,
		b.description,
		enabled,
		webhookConfig,
		subscriptions)
	return config
}
