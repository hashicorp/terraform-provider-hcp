package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var testAccHvnPeeringConnectionConfig = `
resource "hcp_hvn" "test_1" {
	hvn_id         = "test-1"
	cloud_provider = "aws"
	region         = "us-west-2"
	cidr_block     = "172.25.16.0/20"
}

resource "hcp_hvn" "test_2" {
	hvn_id         = "test-2"
	cloud_provider = "aws"
	region         = "us-west-2"
	cidr_block     = "172.18.16.0/20"
}

resource "hcp_hvn_peering_connection" "test" {
	peering_id = "test-peering"
	hvn_1      = hcp_hvn.test_1.self_link
	hvn_2      = hcp_hvn.test_2.self_link
}

data "hcp_hvn_peering_connection" "test" {
	peering_id = "test-peering"
	hvn_1      = hcp_hvn.test_1.self_link
	hvn_2      = hcp_hvn.test_2.self_link
}
`

// This includes tests against both the resource and the corresponding datasource
// to shorten testing time
func TestAccHvnPeeringConnection(t *testing.T) {
	resourceName := "hcp_hvn_peering_connection.test"
	dataSourceName := "data.hcp_hvn_peering_connection.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckHvnPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", "test-peering"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					testLink(resourceName, "hvn_1", "test-1", HvnResourceType, resourceName),
					testLink(resourceName, "hvn_2", "test-2", HvnResourceType, resourceName),
				),
			},
			// Tests import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "test-1:test-peering",
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", "test-peering"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					testLink(resourceName, "hvn_1", "test-1", HvnResourceType, resourceName),
					testLink(resourceName, "hvn_2", "test-2", HvnResourceType, resourceName),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccHvnPeeringConnectionConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peering_id", dataSourceName, "peering_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_1", dataSourceName, "hvn_1"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_2", dataSourceName, "hvn_2"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "expires_at", dataSourceName, "expires_at"),
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

		hvn, err := clients.GetHvnByID(context.Background(), client, hvnLink.Location, hvnLink.ID)
		if err != nil {
			return fmt.Errorf("unable to find hvn for %q: %v", id, err)
		}

		peeringID := peeringLink.ID
		loc := peeringLink.Location

		if _, err := clients.GetPeeringByID(context.Background(), client, peeringID, hvn.ID, loc); err != nil {
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
			hvn, err := clients.GetHvnByID(context.Background(), client, peeringLink.Location, hvnLink.ID)
			if err != nil {
				return fmt.Errorf("unable to find hvn for %q: %v", id, err)
			}

			peeringID := peeringLink.ID
			loc := peeringLink.Location

			_, err = clients.GetPeeringByID(context.Background(), client, peeringID, hvn.ID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed peering connection %q: %v", id, err)
			}
		default:
			continue
		}
	}
	return nil
}
