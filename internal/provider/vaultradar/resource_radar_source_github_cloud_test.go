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

func TestRadarSourceGitHubCloud(t *testing.T) {
	// Requires Project already setup with Radar.
	// Requires a Service Account with an Admin role on the Project.
	// Requires access to a GitHub Cloud Organization.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	githubOrganization := os.Getenv("RADAR_GITHUB_CLOUD_ORGANIZATION")
	token := os.Getenv("RADAR_GITHUB_CLOUD_TOKEN")
	updateToken := os.Getenv("RADAR_GITHUB_CLOUD_TOKEN_2")

	if projectID == "" || githubOrganization == "" || token == "" || updateToken == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_GITHUB_CLOUD_ORGANIZATION, RADAR_GITHUB_CLOUD_TOKEN and " +
			"RADAR_GITHUB_CLOUD_TOKEN_2 must be set for acceptance tests")
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
					resource.TestCheckResourceAttrSet("hcp_vault_radar_source_github_cloud.example", "token"),
				),
			},
			// UPDATE token
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_source_github_cloud" "example" {
						project_id = %q
						github_organization = %q
						token = %q
					}				
				`, projectID, githubOrganization, updateToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("hcp_vault_radar_source_github_cloud.example", "token", func(value string) error {
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
