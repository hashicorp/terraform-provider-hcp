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

func TestRadarIntegrationSlackConnection(t *testing.T) {
	// Requires Project to be with Radar tenant provisioned.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a Slack instance.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	token := os.Getenv("RADAR_INTEGRATION_SLACK_TOKEN")
	updateToken := os.Getenv("RADAR_INTEGRATION_SLACK_TOKEN_2")

	if projectID == "" || token == "" || updateToken == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_INTEGRATION_SLACK_TOKEN and RADAR_INTEGRATION_SLACK_TOKEN_2 " +
			"must be set for acceptance tests")
	}

	name := "AC Test of Creating Slack Connect from TF"
	updatedName := "AC Test of Updating Slack Connect from TF"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_slack_connection" "example" {
						project_id = %q
						name = %q
						token = %q	
					}
				`, projectID, name, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_connection.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_connection.example", "name", name),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_slack_connection.example", "id"),
				),
			},
			// Update name.
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_slack_connection" "example" {
						project_id = %q
						name = %q
						token = %q	
					}
				`, projectID, updatedName, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_connection.example", "name", updatedName),
				),
			},
			// Update token. This effectively cause an update in the auth_key of the connection.
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_slack_connection" "example" {
						project_id = %q
						name = %q
						token = %q	
					}
				`, projectID, updatedName, updateToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("hcp_vault_radar_integration_slack_connection.example", "token", func(value string) error {
						if value != updateToken {
							// Avoid outputting the token in the error message.
							return fmt.Errorf("expected token to be updated")
						}
						return nil
					}),
				),
			},
			// DELETE happens automatically.
		},
	})
}
