// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var boundaryUniqueID = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))

var boundaryClusterResourceTemplate = fmt.Sprintf(`
resource hcp_boundary_cluster "test" {
	cluster_id = "%[1]s"
	username = "test-user"
	password = "password123!"
	tier = "PluS"
	%%s
}
`, boundaryUniqueID)

var maintenanceWindowConfig = `
	maintenance_window_config {
		day   = "TUESDAY"
		start = 2
		end   = 12
		upgrade_type             = "SCHEDULED"
	}
`

var controllerConfig = `
	auth_token_time_to_live = "12h0m0s"
	auth_token_time_to_stale = "1h0m0s"
`

var boundaryCluster = fmt.Sprintf(boundaryClusterResourceTemplate, "")
var boundaryClusterWithMaintenanceWindow = fmt.Sprintf(boundaryClusterResourceTemplate, maintenanceWindowConfig)
var boundaryClusterWithControllerConfig = fmt.Sprintf(boundaryClusterResourceTemplate, controllerConfig)

func setTestAccBoundaryClusterConfig(boundaryCluster string) string {
	return fmt.Sprintf(`
%s

data "hcp_boundary_cluster" "test" {
	cluster_id = hcp_boundary_cluster.test.cluster_id
}
`, boundaryCluster)
}

func TestAcc_Boundary_Cluster(t *testing.T) {
	t.Parallel()

	boundaryClusterResourceName := "hcp_boundary_cluster.test"
	boundaryClusterDataSourceName := "data.hcp_boundary_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBoundaryClusterDestroy,
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
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "tier", "PLUS"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "version"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_live"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_stale"),
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
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "tier", "PLUS"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "version"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_live"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_stale"),
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
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "tier", "PLUS"),
					resource.TestCheckResourceAttrPair(boundaryClusterResourceName, "version", boundaryClusterDataSourceName, "version"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_live"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_stale"),
				),
			},
			{
				// this test step tests creating a boundary cluster maintenance window.
				Config: testConfig(setTestAccBoundaryClusterConfig(boundaryClusterWithMaintenanceWindow)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoundaryClusterExists(boundaryClusterResourceName),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "cluster_id", boundaryUniqueID),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "cluster_url"),
					testAccCheckFullURL(boundaryClusterResourceName, "cluster_url", ""),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "state"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "tier", "PLUS"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "maintenance_window_config.0.upgrade_type", "SCHEDULED"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "maintenance_window_config.0.day", "TUESDAY"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "maintenance_window_config.0.start", "2"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "maintenance_window_config.0.end", "12"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "version"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_live"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "auth_token_time_to_stale"),
				),
			},
			{
				// this test step tests creating a boundary cluster with non-default controller config settings.
				Config: testConfig(setTestAccBoundaryClusterConfig(boundaryClusterWithControllerConfig)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoundaryClusterExists(boundaryClusterResourceName),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "cluster_id", boundaryUniqueID),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "cluster_url"),
					testAccCheckFullURL(boundaryClusterResourceName, "cluster_url", ""),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "state"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "tier", "PLUS"),
					resource.TestCheckResourceAttrSet(boundaryClusterResourceName, "version"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "auth_token_time_to_live", "12h0m0s"),
					resource.TestCheckResourceAttr(boundaryClusterResourceName, "auth_token_time_to_stale", "1h0m0s"),
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
