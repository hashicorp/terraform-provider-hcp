package provider

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	// using unique names for AWS resource to make debugging easier
	hvnRouteUniqueAWSName = fmt.Sprintf("hcp-tf-provider-test-%d", rand.Intn(99999))
	testAccHvnRouteConfig = fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.220.0.0/16"
  tags = {
     Name = "%[1]s"
  }
}

resource "hcp_aws_network_peering" "peering" {	
  peering_id      = "hcp-tf-provider-test"
  hvn_id          = hcp_hvn.test.hvn_id
  peer_account_id = aws_vpc.vpc.owner_id
  peer_vpc_id     = aws_vpc.vpc.id
  peer_vpc_region = "us-west-2"
}

resource "hcp_hvn_route" "route" {
  hvn_route_id = "peering-route"
  hvn_link = hcp_hvn.test.self_link
  destination_cidr = "172.31.0.0/16"
  target_link = hcp_aws_network_peering.peering.self_link
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
`, hvnRouteUniqueAWSName)
)

func TestAccHvnRoute(t *testing.T) {
	resourceName := "hcp_hvn_route.route"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, true) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {VersionConstraint: "~> 2.64.0"},
		},
		CheckDestroy: testAccCheckHvnRouteDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply created correct HVN route
			{
				Config: testConfig(testAccHvnRouteConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", "peering-route"),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					testLink(resourceName, "self_link", "peering-route", HVNRouteResourceType, "hcp_hvn.test"),
					testLink(resourceName, "target_link", "hcp-tf-provider-test", PeeringResourceType, "hcp_hvn.test"),
				),
			},
			// Testing that we can import HVN route created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "test-hvn:peering-route",
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccHvnRouteConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnRouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_route_id", "peering-route"),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr", "172.31.0.0/16"),
					testLink(resourceName, "self_link", "peering-route", HVNRouteResourceType, "hcp_hvn.test"),
					testLink(resourceName, "target_link", "hcp-tf-provider-test", PeeringResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func testAccCheckHvnRouteExists(name string) resource.TestCheckFunc {
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

		hvnRouteLink, err := buildLinkFromURL(id, HVNRouteResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build hvnRouteLink for %q: %v", id, err)
		}

		hvnUrl, ok := rs.Primary.Attributes["hvn_link"]
		if !ok {
			return fmt.Errorf("hcp_hvn_route doesn't have hvn_link")
		}
		hvnLink, err := parseLinkURL(hvnUrl, HvnResourceType)
		if err != nil {
			return fmt.Errorf("failed to parse hvn_link: %w", err)
		}

		hvnRouteID := hvnRouteLink.ID
		loc := hvnRouteLink.Location

		if _, err := clients.GetHVNRoute(context.Background(), client, hvnLink.ID, hvnRouteID, loc); err != nil {
			return fmt.Errorf("unable to get HVN route %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckHvnRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_hvn_route":
			id := rs.Primary.ID

			hvnRouteLink, err := buildLinkFromURL(id, HVNRouteResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build hvnRouteLink for %q: %v", id, err)
			}

			hvnUrl, ok := rs.Primary.Attributes["hvn_link"]
			if !ok {
				return fmt.Errorf("hcp_hvn_route doesn't have hvn_link")
			}
			hvnLink, err := parseLinkURL(hvnUrl, HvnResourceType)
			if err != nil {
				return fmt.Errorf("failed to parse hvn_link: %w", err)
			}

			hvnRouteID := hvnRouteLink.ID
			loc := hvnRouteLink.Location

			_, err = clients.GetHVNRoute(context.Background(), client, hvnLink.ID, hvnRouteID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
