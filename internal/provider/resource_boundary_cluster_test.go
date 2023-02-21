// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var boundaryUniqueID = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))

var boundaryCluster = fmt.Sprintf(`
resource hcp_boundary_cluster "test" {
	cluster_id = "%[1]s"
	username = "test-user"
	password = "password123!"
}
`, boundaryUniqueID)

func setTestAccBoundaryClusterConfig(boundaryCluster string) string {
	return fmt.Sprintf(`
%s

data "hcp_boundary_cluster" "test" {
	cluster_id = hcp_boundary_cluster.test.cluster_id
}
`, boundaryCluster)
}

func TestAccBoundaryCluster(t *testing.T) {
	boundaryClusterResourceName := "hcp_boundary_cluster.test"
	boundaryClusterDataSourceName := "data.hcp_boundary_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckBoundaryClusterDestroy,
		Steps: []resource.TestStep{
			{
				// this test step tests boundary cluster creation.
				Config: testConfig(setTestAccBoundaryClusterConfig(boundaryCluster)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoundaryClusterExists(boundaryClusterResourceName),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "cluster_id", boundaryUniqueID),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "cluster_url"),
					testAccCheckFullURL(boundaryClusterResourceName, "cluster_url", ""),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "state"),
				),
			},
			{
				// this test step simulates the import of a boundary cluster.
				ResourceName: boundaryClusterResourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[boundaryClusterResourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", boundaryClusterResourceName)
					}

					return rs.Primary.Attributes["cluster_id"], nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"username", "password"},
			},
			{
				// this test step is a subsequent terraform apply that verifies no state is modified.
				Config: testConfig(setTestAccBoundaryClusterConfig(boundaryCluster)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoundaryClusterExists(boundaryClusterResourceName),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "cluster_id", boundaryUniqueID),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "cluster_url"),
					testAccCheckFullURL(boundaryClusterResourceName, "cluster_url", ""),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "state"),
				),
			},
			{
				// this step tests the data source.
				Config: testConfig(setTestAccBoundaryClusterConfig(boundaryCluster)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(boundaryClusterResourceName, "cluster_id", boundaryClusterDataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(boundaryClusterResourceName, "created_at", boundaryClusterDataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(boundaryClusterResourceName, "cluster_url", boundaryClusterDataSourceName, "cluster_url"),
					testAccCheckFullURL(boundaryClusterDataSourceName, "cluster_url", ""),
					resource.TestCheckResourceAttrPair(boundaryClusterResourceName, "state", boundaryClusterDataSourceName, "state"),
				),
			},
		},
	})
}

func testAccCheckBoundaryClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_boundary_cluster":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, BoundaryClusterResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			clusterID := link.ID
			loc := link.Location

			_, err = clients.GetBoundaryClusterByID(context.Background(), client, loc, clusterID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Boundary cluster %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}

func testAccCheckBoundaryClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*clients.Client)

		link, err := buildLinkFromURL(id, BoundaryClusterResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		clusterID := link.ID
		loc := link.Location

		if _, err := clients.GetBoundaryClusterByID(context.Background(), client, loc, clusterID); err != nil {
			return fmt.Errorf("unable to read Boundary cluster %q: %v", id, err)
		}

		return nil
	}
}
