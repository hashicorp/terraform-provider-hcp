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

func TestRadarResources(t *testing.T) {
	// Requires Radar project already setup
	// Requires least one resource set up and registered with HCP.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	resourceURI := os.Getenv("RADAR_RESOURCES_URI_LIKE_FILTER")

	if projectID == "" || resourceURI == "" {
		t.Skip("HCP_PROJECT_ID and RADAR_RESOURCES_URI_LIKE_FILTER must be set for acceptance tests")
	}

	radarResourcesName := "data.hcp_vault_radar_resources.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read the resources with a URI filter for the specific resource.
			{
				Config: createRadarResourcesConfig(projectID, resourceURI),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(radarResourcesName, "project_id", projectID),
					resource.TestCheckResourceAttr(radarResourcesName, "uri_like_filter.values.0", resourceURI),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.data_source_name"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.data_source_type"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.detector_type"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.hcp_resource_name"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.hcp_resource_status"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.id"),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.name"),
					resource.TestCheckResourceAttr(radarResourcesName, "resources.0.state", "created"),
					resource.TestCheckResourceAttr(radarResourcesName, "resources.0.uri", resourceURI),
					resource.TestCheckResourceAttrSet(radarResourcesName, "resources.0.visibility"),
				),
			},
		},
	})
}

func createRadarResourcesConfig(projectID, resourceURI string) string {
	return fmt.Sprintf(`
		data hcp_vault_radar_resources "example" {
          project_id = %q
		  uri_like_filter = {
			values = [%q]
			case_insensitive = false
          }
		}
		`, projectID, resourceURI)
}
