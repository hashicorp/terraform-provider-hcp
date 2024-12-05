// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestRadarIntegrationJiraConnection(t *testing.T) {
	// Requires Project to be with Radar tenant provisioned.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a Jira instance.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	baseURL := os.Getenv("RADAR_INTEGRATION_JIRA_BASE_URL")
	email := os.Getenv("RADAR_INTEGRATION_JIRA_EMAIL")
	token := os.Getenv("RADAR_INTEGRATION_JIRA_TOKEN")

	if projectID == "" || baseURL == "" || email == "" || token == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_INTEGRATION_JIRA_BASE_URL, RADAR_INTEGRATION_JIRA_EMAIL, and RADAR_INTEGRATION_JIRA_TOKEN must be set for acceptance tests")
	}

	name := "AC Test of Creating Jira Connect from TF"
	updatedName := "AC Test of UpdatingJira Connect from TF"

	// To simulate a change in the base_url, we will split the base_url and update the first part to be uppercase.
	// Example: https://acme.atlassian.net will become https://ACME.atlassian.net
	protocolAndURI := strings.SplitN(baseURL, "://", 2)
	protocol := protocolAndURI[0]
	uri := protocolAndURI[1]
	parts := strings.SplitN(uri, ".", 2)
	updatedBaseURL := fmt.Sprintf("%s://%s", protocol, strings.ToUpper(parts[0])+"."+parts[1])

	// To simulate a change in the email, we will update the email to be uppercase.
	updateEmail := strings.ToUpper(email)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_jira_connection" "example" {
						project_id = %q
						name = %q
						base_url = %q
						email = %q
						token = %q	
					}
				`, projectID, name, baseURL, email, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "name", name),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "base_url", baseURL),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "email", email),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_jira_connection.example", "id"),
				),
			},
			// Update name.
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_jira_connection" "example" {
						project_id = %q
						name = %q
						base_url = %q
						email = %q
						token = %q	
					}
				`, projectID, updatedName, baseURL, email, token),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_jira_connection.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "name", updatedName),
				),
			},
			// Update base_url. This effectively cause an update in the details of the connection.
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_jira_connection" "example" {
						project_id = %q
						name = %q
						base_url = %q
						email = %q
						token = %q	
					}
				`, projectID, updatedName, updatedBaseURL, email, token),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_jira_connection.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "base_url", updatedBaseURL),
				),
			},
			// Update email. This effectively cause an update in the auth_key of the connection.
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_jira_connection" "example" {
						project_id = %q
						name = %q
						base_url = %q
						email = %q
						token = %q	
					}
				`, projectID, updatedName, updatedBaseURL, updateEmail, token),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_integration_jira_connection.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "email", updateEmail),
				),
			},
			// DELETE happens automatically.
		},
	})
}
