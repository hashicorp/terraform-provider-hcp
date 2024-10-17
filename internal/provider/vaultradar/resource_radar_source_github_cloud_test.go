package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestRadarSourceGitHubCloud(t *testing.T) {
	// Requires Project already setup with Radar.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a GitHub Cloud Organization.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	githubOrganization := os.Getenv("RADAR_GITHUB_CLOUD_ORGANIZATION")
	token := os.Getenv("RADAR_GITHUB_CLOUD_TOKEN")

	if projectID == "" || githubOrganization == "" || token == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_GITHUB_CLOUD_ORGANIZATION, and RADAR_GITHUB_CLOUD_TOKEN must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_source_github_cloud" "example" {
						project_id = %q
						github_organization = %q
						token = %q
					}				
				`, projectID, githubOrganization, token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_source_github_cloud.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_source_github_cloud.example", "github_organization", githubOrganization),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_source_github_cloud.example", "id"),
				),
			},
			// UPDATE not supported at this time.
			// DELETE happens automatically.
		},
	})
}
