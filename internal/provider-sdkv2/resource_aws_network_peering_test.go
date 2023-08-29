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
	// using unique names for AWS resource to make debugging easier
	hvnPeeringUniqueAWSName = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))
	testAccAwsPeeringConfig = fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.220.0.0/16"
  tags = {
     Name = "%[1]s"
  }
}

// This resource initially returns in a Pending state, because its provider_peering_id is required to complete acceptance of the connection.
resource "hcp_aws_network_peering" "peering" {	
  peering_id                = "%[1]s"
  hvn_id                    = hcp_hvn.test.hvn_id
  peer_account_id           = aws_vpc.vpc.owner_id
  peer_vpc_id               = aws_vpc.vpc.id
  peer_vpc_region           = "us-west-2"
}

// This data source is the same as the resource above, but waits for the connection to be Active before returning.
data "hcp_aws_network_peering" "peering" {
  hvn_id                    = hcp_hvn.test.hvn_id
  peering_id                = hcp_aws_network_peering.peering.peering_id
  wait_for_active_state     = true
}

// The route depends on the data source, rather than the resource, to ensure the peering is in an Active state.
resource "hcp_hvn_route" "route" {
  hvn_route_id              = "%[1]s"
  hvn_link                  = hcp_hvn.test.self_link
  destination_cidr          = "172.31.0.0/16"
  target_link               = data.hcp_aws_network_peering.peering.self_link
}

resource "aws_vpc_peering_connection_accepter" "peering-accepter" {
  vpc_peering_connection_id = hcp_aws_network_peering.peering.provider_peering_id
  auto_accept               = true
  tags = {
     Name = "%[1]s"

	 // we need to have these tags here because peering-accepter will turn into
     // an actual peering which HCP will populate with a set of tags (the ones below).
     // After succesfull "apply"" test will try to run "plan" operation
     // to make sure there are no changes to the state and if we don't specify these
     // tags here then it will fail. 
	 hvn_id          = hcp_hvn.test.hvn_id
	 organization_id = hcp_hvn.test.organization_id
	 project_id      = hcp_hvn.test.project_id
	 peering_id      = hcp_aws_network_peering.peering.peering_id
  }
}
`, hvnPeeringUniqueAWSName)
)

func TestAccAwsPeering(t *testing.T) {
	resourceName := "hcp_aws_network_peering.peering"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": true, "azure": false}) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {VersionConstraint: "~> 4.0.0"},
		},
		CheckDestroy: testAccCheckHvnPeeringDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccAwsPeeringConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", hvnPeeringUniqueAWSName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnPeeringUniqueAWSName),
					resource.TestCheckResourceAttrSet(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vpc_region"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnPeeringUniqueAWSName, PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := s.RootModule().Resources["hcp_hvn.test"].Primary.Attributes["hvn_id"]
					peerID := rs.Primary.Attributes["peering_id"]
					return fmt.Sprintf("%s:%s", hvnID, peerID), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccAwsPeeringConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnPeeringExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "peering_id", hvnPeeringUniqueAWSName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnPeeringUniqueAWSName),
					resource.TestCheckResourceAttrSet(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "peer_vpc_region"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_peering_id"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnPeeringUniqueAWSName, PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func testAccCheckHvnPeeringExists(name string) resource.TestCheckFunc {
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

		hvnID, ok := rs.Primary.Attributes["hvn_id"]
		if !ok {
			return fmt.Errorf("no hvn_id is set")
		}

		peeringID := peeringLink.ID
		loc := peeringLink.Location

		if _, err := clients.GetPeeringByID(context.Background(), client, peeringID, hvnID, loc); err != nil {
			return fmt.Errorf("unable to get TGW attachment %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckHvnPeeringDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_aws_network_peering":
			id := rs.Primary.ID

			if id == "" {
				return fmt.Errorf("no ID is set")
			}

			peeringLink, err := buildLinkFromURL(id, PeeringResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build peeringLink for %q: %v", id, err)
			}

			hvnID, ok := rs.Primary.Attributes["hvn_id"]
			if !ok {
				return fmt.Errorf("no hvn_id is set")
			}

			peeringID := peeringLink.ID
			loc := peeringLink.Location

			_, err = clients.GetPeeringByID(context.Background(), client, peeringID, hvnID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
