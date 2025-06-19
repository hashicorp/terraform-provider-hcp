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

func TestRadarResourceList(t *testing.T) {
	// Requires Radar project already setup
	// Requires least one resource set up and registered with HCP.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	resourceURI := os.Getenv("RADAR_RESOURCE_LIST_URI_LIKE_FILTER")

	if projectID == "" || resourceURI == "" {
		t.Skip("HCP_PROJECT_ID and RADAR_RESOURCE_LIST_URI_LIKE_FILTER must be set for acceptance tests")
	}

	resourceListName := "data.hcp_vault_radar_resource_list.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read the resource list with a URI filter for the specific resource.
			{
				Config: createRadarResourceListConfig(projectID, resourceURI),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceListName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceListName, "uri_like_filter.0", resourceURI),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.data_source_info"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.data_source_name"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.data_source_type"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.detector_type"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.hcp_resource_id"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.hcp_resource_name"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.hcp_resource_status"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.id"),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.name"),
					resource.TestCheckResourceAttr(resourceListName, "resources.0.state", "created"),
					resource.TestCheckResourceAttr(resourceListName, "resources.0.uri", resourceURI),
					resource.TestCheckResourceAttrSet(resourceListName, "resources.0.visibility"),
				),
			},
		},
	})
}

func createRadarResourceListConfig(projectID, resourceURI string) string {
	return fmt.Sprintf(`
		data hcp_vault_radar_resource_list "example" {
          project_id = %q
		  uri_like_filter = [%q]
		}
		`, projectID, resourceURI)
}
