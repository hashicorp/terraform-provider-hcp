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
	tgwAttUniqueAWSName        = fmt.Sprintf("hcp-att-unique-test-%s", time.Now().Format("200601021504"))
	tgwAttUniqueHvnName        = fmt.Sprintf("att-hvn-name-%s", time.Now().Format("200601021504"))
	testAccTGWAttachmentConfig = fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "test" {
  hvn_id         = "%[2]s"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "aws_vpc" "example" {
  cidr_block = "172.31.0.0/16"
  tags = {
    Name = "%[1]s"
  }
}

resource "aws_ec2_transit_gateway" "example" {
  tags = {
    Name = "%[1]s"
  }
}

resource "aws_ram_resource_share" "example" {
  name                      = "%[1]s"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  principal          = hcp_hvn.test.provider_account_id
}

resource "aws_ram_resource_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  resource_arn       = aws_ec2_transit_gateway.example.arn
}

resource "hcp_aws_transit_gateway_attachment" "example" {
  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]

  hvn_id                        = hcp_hvn.test.hvn_id
  transit_gateway_attachment_id = "%[1]s"
  transit_gateway_id            = aws_ec2_transit_gateway.example.id
  resource_share_arn            = aws_ram_resource_share.example.arn
}

// This data source is the same as the resource above, but waits for the connection to be Active before returning.
data "hcp_aws_transit_gateway_attachment" "example" {
  hvn_id                               = hcp_hvn.test.hvn_id
  transit_gateway_attachment_id        = hcp_aws_transit_gateway_attachment.example.transit_gateway_attachment_id
  wait_for_active_state                = true
}

// The route depends on the data source, rather than the resource, to ensure the TGW is in an Active state.
resource "hcp_hvn_route" "route" {
 hvn_link         = hcp_hvn.test.self_link
 hvn_route_id     = "%[1]s"
 destination_cidr = aws_vpc.example.cidr_block
 target_link      = data.hcp_aws_transit_gateway_attachment.example.self_link
}

resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = hcp_aws_transit_gateway_attachment.example.provider_transit_gateway_attachment_id

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
	 tgw_attachment_id = hcp_aws_transit_gateway_attachment.example.transit_gateway_attachment_id
  }
}
`, tgwAttUniqueAWSName, tgwAttUniqueHvnName)
)

func TestAccTGWAttachment(t *testing.T) {
	resourceName := "hcp_aws_transit_gateway_attachment.example"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": true, "azure": false}) },
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {VersionConstraint: "~> 4.0.0"},
		},
		CheckDestroy: testAccCheckTGWAttachmentDestroy,

		Steps: []resource.TestStep{
			// Testing that initial Apply creates correct TGW attachment
			{
				Config: testConfig(testAccTGWAttachmentConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTGWAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", tgwAttUniqueAWSName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", tgwAttUniqueHvnName),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					testLink(resourceName, "self_link", tgwAttUniqueAWSName, TgwAttachmentResourceType, "hcp_hvn.test"),
				),
			},
			// Testing that we can import TGW attachment created in the previous step and that the
			// resource terraform state will be exactly the same
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					resourceShare, ok := s.RootModule().Resources["aws_ram_resource_share.example"]
					if !ok {
						return "", fmt.Errorf("not found: aws_ram_resource_share.example")
					}

					return fmt.Sprintf("%s:%s:%s", tgwAttUniqueHvnName, tgwAttUniqueAWSName, resourceShare.Primary.Attributes["arn"]), nil
				},
				ImportStateVerify: true,
			},
			// Testing running Terraform Apply for already known resource
			{
				Config: testConfig(testAccTGWAttachmentConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTGWAttachmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", tgwAttUniqueAWSName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", tgwAttUniqueHvnName),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					testLink(resourceName, "self_link", tgwAttUniqueAWSName, TgwAttachmentResourceType, "hcp_hvn.test"),
				),
			},
		},
	})
}

func testAccCheckTGWAttachmentExists(name string) resource.TestCheckFunc {
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

		tgwAttLink, err := buildLinkFromURL(id, TgwAttachmentResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build tgwAttLink for %q: %v", id, err)
		}

		hvnID, ok := rs.Primary.Attributes["hvn_id"]
		if !ok {
			return fmt.Errorf("no hvn_id is set")
		}

		tgwAttID := tgwAttLink.ID
		loc := tgwAttLink.Location

		if _, err := clients.GetTGWAttachmentByID(context.Background(), client, tgwAttID, hvnID, loc); err != nil {
			return fmt.Errorf("unable to get TGW attachment %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckTGWAttachmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_aws_transit_gateway_attachment":
			id := rs.Primary.ID
			if id == "" {
				return fmt.Errorf("no ID is set")
			}

			tgwAttLink, err := buildLinkFromURL(id, TgwAttachmentResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build tgwAttLink for %q: %v", id, err)
			}

			hvnID, ok := rs.Primary.Attributes["hvn_id"]
			if !ok {
				return fmt.Errorf("no hvn_id is set")
			}

			tgwAttID := tgwAttLink.ID
			loc := tgwAttLink.Location

			_, err = clients.GetTGWAttachmentByID(context.Background(), client, tgwAttID, hvnID, loc)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
