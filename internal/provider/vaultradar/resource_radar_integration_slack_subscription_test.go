// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

	name := "AC Test of Creating Slack Subscription From TF"
	updatedName := "AC Test of Updating Slack Subscription From TF"

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
						name = %q
						connection_id = hcp_vault_radar_integration_slack_connection.slack_connection.id
						channel = %q
					}
						
				`, projectID, token, name, channel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_slack_subscription.slack_subscription", "connection_id"),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_subscription.slack_subscription", "name", name),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_subscription.slack_subscription", "channel", channel),
				),
			},
			// UPDATE name.
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
						name = %q
						connection_id = hcp_vault_radar_integration_slack_connection.slack_connection.id
						channel = %q
					}
						
				`, projectID, token, updatedName, channel),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_slack_connection.slack_connection", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_slack_subscription.slack_subscription", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_subscription.slack_subscription", "name", updatedName),
				),
			},
			// UPDATE Connection ID.
			{
				Config: fmt.Sprintf(`
					# An integration_slack_subscription is required to create a hcp_vault_radar_integration_slack_subscription.
					resource "hcp_vault_radar_integration_slack_connection" "slack_connection" {
						project_id = %q
						name = "AC Test of Slack Connect from TF"
						token = %q	
					}

					# Create another integration_slack_subscription to connect to.
					resource "hcp_vault_radar_integration_slack_connection" "slack_connection_2" {
						project_id = %q
						name = "AC Test of Slack Connect from TF 2"
						token = %q	
					}

					resource "hcp_vault_radar_integration_slack_subscription" "slack_subscription" {
						project_id = hcp_vault_radar_integration_slack_connection.slack_connection.project_id
						name = %q
						connection_id = hcp_vault_radar_integration_slack_connection.slack_connection_2.id
						channel = %q
					}
						
				`, projectID, token, projectID, token, updatedName, channel),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_slack_connection.slack_connection", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_slack_connection.slack_connection_2", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_slack_subscription.slack_subscription", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_slack_subscription.slack_subscription", "connection_id"),
				),
			},
			// DELETE happens automatically.
		},
	})
}
