package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	uniqueActivationTestID         = fmt.Sprintf("hcp-tf-provider-test-%d", rand.Intn(99999))
	activationHvnID                = uniqueActivationTestID + "-hvn"
	activationPeeringID            = uniqueActivationTestID + "-peering"
	testAccPeeringActivationConfig = fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "test" {
	hvn_id         = "%[2]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.220.0.0/16"
  tags = {
     Name = "hcp-provider-peering-activation-test"
  }
}

resource "hcp_aws_network_peering" "peering" {	
  peering_id      = "%[3]s"
  hvn_id          = hcp_hvn.test.hvn_id
  peer_account_id = aws_vpc.vpc.owner_id
  peer_vpc_id     = aws_vpc.vpc.id
  peer_vpc_region = "us-west-2"
}

resource "hcp_peering_connection_activation" "activation" {
  peering_id  = hcp_aws_network_peering.peering.peering_id
  hvn_link    = hcp_hvn.test.self_link
}

resource "hcp_hvn_route" "route" {
  hvn_route_id      = "%[1]s-route"
  hvn_link          = hcp_hvn.test.self_link
  destination_cidr  = "172.31.0.0/16"
  target_link       = hcp_aws_network_peering.peering.self_link
}

resource "aws_vpc_peering_connection_accepter" "peering-accepter" {
  vpc_peering_connection_id = hcp_aws_network_peering.peering.provider_peering_id
  auto_accept               = true
  tags = {
    Name            = "hcp-provider-peering-activation-test"
    hvn_id          = hcp_hvn.test.hvn_id
    organization_id = hcp_hvn.test.organization_id
    project_id      = hcp_hvn.test.project_id
    peering_id      = hcp_aws_network_peering.peering.peering_id
  }
}
`, uniqueActivationTestID, activationHvnID, activationPeeringID)
)

func TestAccPeeringConnectionActivation(t *testing.T) {
	resourceName := "hcp_peering_connection_activation.activation"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, true) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {VersionConstraint: "~> 2.64.0"},
		},

		Steps: []resource.TestStep{
			{
				// Tests create
				Config: testConfig(testAccPeeringActivationConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "peering_id", activationPeeringID),
					testLink(resourceName, "hvn_link", activationHvnID, HvnResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}
