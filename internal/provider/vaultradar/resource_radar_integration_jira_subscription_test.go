// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestRadarIntegrationJiraSubscription(t *testing.T) {
	// Requires Project to be with Radar tenant provisioned.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a Jira instance.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	baseURL := os.Getenv("RADAR_INTEGRATION_JIRA_BASE_URL")
	email := os.Getenv("RADAR_INTEGRATION_JIRA_EMAIL")
	token := os.Getenv("RADAR_INTEGRATION_JIRA_TOKEN")
	jiraProjectKey := os.Getenv("RADAR_INTEGRATION_JIRA_PROJECT_KEY")
	issueType := os.Getenv("RADAR_INTEGRATION_JIRA_ISSUE_TYPE")
	assignee := os.Getenv("RADAR_INTEGRATION_JIRA_ASSIGNEE")

	// For the connection resource.
	if projectID == "" || baseURL == "" || email == "" || token == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_INTEGRATION_JIRA_BASE_URL, RADAR_INTEGRATION_JIRA_EMAIL, and RADAR_INTEGRATION_JIRA_TOKEN must be set for acceptance tests")
	}

	// For the subscription resource.
	if jiraProjectKey == "" || issueType == "" || assignee == "" {
		t.Skip("RADAR_INTEGRATION_JIRA_PROJECT_KEY, RADAR_INTEGRATION_JIRA_ISSUE_TYPE, and RADAR_INTEGRATION_JIRA_ASSIGNEE must be set for acceptance tests")
	}

	message := "AC test message"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					# An integration_jira_subscription is required to create a hcp_vault_radar_integration_jira_subscription.
					resource "hcp_vault_radar_integration_jira_connection" "jira_connection" {
						project_id = %q
						name = "AC Test of Jira Connect from TF"
						base_url = %q
						email = %q
						token = %q	
					}

					resource "hcp_vault_radar_integration_jira_subscription" "jira_subscription" {
						project_id = hcp_vault_radar_integration_jira_connection.jira_connection.project_id
						name = "AC Test of Jira Subscription From TF"
						connection_id = hcp_vault_radar_integration_jira_connection.jira_connection.id
						jira_project_key = %q
						issue_type = %q
						assignee = %q
						message = %q
					}
						
				`, projectID, baseURL, email, token,
					jiraProjectKey, issueType, assignee, message),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_jira_subscription.jira_subscription", "connection_id"),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_subscription.jira_subscription", "jira_project_key", jiraProjectKey),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_subscription.jira_subscription", "issue_type", issueType),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_subscription.jira_subscription", "assignee", assignee),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_subscription.jira_subscription", "message", message),
				),
			},
			// UPDATE not supported at this time.
			// DELETE happens automatically.
		},
	})
}
