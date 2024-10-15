package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestRadarIntegrationSlackSubscription(t *testing.T) {
	// Requires Project to be with Radar tenant provisioned.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a Slack instance.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	token := os.Getenv("RADAR_INTEGRATION_SLACK_TOKEN")
	channel := os.Getenv("RADAR_INTEGRATION_SLACK_CHANNEL")

	if projectID == "" || token == "" || channel == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_INTEGRATION_SLACK_TOKEN and RADAR_INTEGRATION_SLACK_CHANNEL must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					# An integration_slack_subscription is required to create a hcp_vault_radar_integration_slack_subscription.
					resource "hcp_vault_radar_integration_slack_connection" "slack_connection" {
						project_id = %q
						name = "AC Test of Slack Connect from TF"
						token = %q	
					}

					resource "hcp_vault_radar_integration_slack_subscription" "slack_subscription" {
						project_id = hcp_vault_radar_integration_slack_connection.slack_connection.project_id
						name = "AC Test of Slack Subscription From TF"
						connection_id = hcp_vault_radar_integration_slack_connection.slack_connection.id
						channel = %q
					}
						
				`, projectID, token, channel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_slack_subscription.slack_subscription", "connection_id"),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_subscription.slack_subscription", "channel", channel),
				),
			},
			// UPDATE not supported at this time.
			// DELETE happens automatically.
		},
	})
}
