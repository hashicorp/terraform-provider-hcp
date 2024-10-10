package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_integration_jira_connection" "example" {
						project_id = %q
						name = "AC Test of Jira Connect from TF"
						base_url = %q
						email = %q
						token = %q	
					}
				`, projectID, baseURL, email, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "base_url", baseURL),
					resource.TestCheckResourceAttr("hcp_vault_radar_integration_jira_connection.example", "email", email),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_integration_jira_connection.example", "id"),
				),
			},
			// UPDATE not supported at this time.
			// DELETE happens automatically.
		},
	})
}
