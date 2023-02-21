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

var (
	hvn1UniqueID = fmt.Sprintf("hcp-provider-test-%s-1", time.Now().Format("200601021504"))
	hvn2UniqueID = fmt.Sprintf("hcp-provider-test-%s-2", time.Now().Format("200601021504"))
)

var testAccHvnPeeringConnectionConfig = fmt.Sprintf(`
resource "hcp_hvn" "test_1" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
	cidr_block     = "172.25.16.0/20"
}

resource "hcp_hvn" "test_2" {
	hvn_id         = "%[2]s"
	cloud_provider = "aws"
	region         = "us-west-2"
	cidr_block     = "172.18.16.0/20"
}

resource "hcp_hvn_peering_connection" "test" {
	hvn_1      = hcp_hvn.test_1.self_link
	hvn_2      = hcp_hvn.test_2.self_link
}

data "hcp_hvn_peering_connection" "test" {
	peering_id = hcp_hvn_peering_connection.test.peering_id
	hvn_1      = hcp_hvn_peering_connection.test.hvn_1
	hvn_2      = hcp_hvn_peering_connection.test.hvn_2
}
`, hvn1UniqueID, hvn2UniqueID)

// This includes tests against both the resource and the corresponding datasource
// to shorten testing time
func TestAccHvnPeeringConnection(t *testing.T) {
	resourceName := "hcp_hvn_peering_connection.test"
	dataSourceName := "data.hcp_hvn_peering_connection.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckHvnPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringConnectionExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "hvn_1", hvn1UniqueID, HvnResourceType, resourceName),
					testLink(resourceName, "hvn_2", hvn2UniqueID, HvnResourceType, resourceName),
				),
			},
			// Tests import
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test_1"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringConnectionExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "hvn_1", hvn1UniqueID, HvnResourceType, resourceName),
					testLink(resourceName, "hvn_2", hvn2UniqueID, HvnResourceType, resourceName),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peering_id", dataSourceName, "peering_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "expires_at", dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
					testLink(dataSourceName, "hvn_1", hvn1UniqueID, HvnResourceType, dataSourceName),
					testLink(dataSourceName, "hvn_2", hvn2UniqueID, HvnResourceType, dataSourceName),
				),
			},
		},
	})
}

func testAccCheckHvnPeeringConnectionExists(name string) resource.TestCheckFunc {
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

		peeringLink, err := buildLinkFromURL(id, PeeringResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build peeringLink for %q: %v", id, err)
		}

		hvnSelfLink, ok := rs.Primary.Attributes["hvn_1"]
		if !ok {
			return fmt.Errorf("no hvn_1 is set")
		}

		hvnLink, err := buildLinkFromURL(hvnSelfLink, HvnResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to parse hvn_1 link URL for %q: %v", id, err)
		}

		peeringID := peeringLink.ID
		loc := peeringLink.Location

		if _, err := clients.GetPeeringByID(context.Background(), client, peeringID, hvnLink.ID, loc); err != nil {
			return fmt.Errorf("unable to get peering connection %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckHvnPeeringConnectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_hvn_peering_connection":
			id := rs.Primary.ID
			if id == "" {
				return fmt.Errorf("no ID is set")
			}

			peeringLink, err := buildLinkFromURL(id, PeeringResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build peeringLink for %q: %v", id, err)
			}

			hvnLink, err := buildLinkFromURL(rs.Primary.Attributes["hvn_1"], HvnResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to parse hvn_1 link URL for %q: %v", id, err)
			}

			peeringID := peeringLink.ID
			loc := peeringLink.Location

			_, err = clients.GetPeeringByID(context.Background(), client, peeringID, hvnLink.ID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed peering connection %q: %v", id, err)
			}
		default:
			continue
		}
	}
	return nil
}
