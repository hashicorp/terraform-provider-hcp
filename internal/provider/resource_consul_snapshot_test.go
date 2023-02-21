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

var consulClusterSnapshotUniqueID = fmt.Sprintf("test-snapshot-%s", time.Now().Format("200601021504"))
var consulClusterSnapshotHVNUniqueID = fmt.Sprintf("test-snapshot-hvn-%s", time.Now().Format("200601021504"))

var testAccConsulSnapshotConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "hcp_consul_cluster" "test" {
	cluster_id = "%[2]s"
	hvn_id     = hcp_hvn.test.hvn_id
	tier       = "development"
}

resource "hcp_consul_snapshot" "test" {
	cluster_id    = hcp_consul_cluster.test.cluster_id
	snapshot_name = "%[2]s"
}`, consulClusterSnapshotHVNUniqueID, consulClusterSnapshotUniqueID)

func TestAccConsulSnapshot(t *testing.T) {
	resourceName := "hcp_consul_snapshot.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckConsulSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(testAccConsulSnapshotConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", consulClusterSnapshotUniqueID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", consulClusterSnapshotUniqueID),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_version"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckNoResourceAttr(resourceName, "restored_at"), // Not a restored snapshot
				),
			},
			{
				Config: testConfig(testAccConsulSnapshotConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulSnapshotExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", consulClusterSnapshotUniqueID),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", consulClusterSnapshotUniqueID),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_version"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckNoResourceAttr(resourceName, "restored_at"), // Not a restored snapshot
				),
			},
		},
	})
}

func testAccCheckConsulSnapshotExists(name string) resource.TestCheckFunc {
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

		link, err := buildLinkFromURL(id, ConsulSnapshotResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		snapshotID := link.ID
		loc := link.Location

		if _, err := clients.GetSnapshotByID(context.Background(), client, loc, snapshotID); err != nil {
			return fmt.Errorf("unable to read Consul snapshot %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckConsulSnapshotDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_consul_snapshot":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, ConsulSnapshotResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			snapshotID := link.ID
			loc := link.Location

			_, err = clients.GetSnapshotByID(context.Background(), client, loc, snapshotID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Consul snapshot %q: %v", id, err)
			}
		default:
			continue
		}
	}

	return nil
}
