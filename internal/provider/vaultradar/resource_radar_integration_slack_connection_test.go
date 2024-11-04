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

	if projectID == "" || token == "" {
		t.Skip("HCP_PROJECT_ID and RADAR_INTEGRATION_SLACK_TOKEN must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_slack_connection" "example" {
						project_id = %q
						name = "AC Test of Slack Connect from TF"
						token = %q	
					}
				`, projectID, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_slack_connection.example", "project_id", projectID),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_slack_connection.example", "id"),
				),
			},
			// UPDATE not supported at this time.
			// DELETE happens automatically.
		},
	})
}
