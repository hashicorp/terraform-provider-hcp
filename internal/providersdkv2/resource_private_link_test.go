// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func TestAccResourcePrivateLink(t *testing.T) {
	resourceName := "hcp_private_link.test"
	hvnResourceName := "hcp_hvn.test"
	vaultResourceName := "hcp_vault_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": true}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPrivateLinkDestroy,
		Steps: []resource.TestStep{
			// Initial creation
			{
				Config: testAccResourcePrivateLinkBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "private_link_id", "test-private-link"),
					resource.TestCheckResourceAttr(resourceName, "consumer_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "consumer_regions.0", "us-west-1"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", hvnResourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_cluster_id", vaultResourceName, "cluster_id"),
					resource.TestCheckResourceAttr(resourceName, "default_region", "us-west-2"),
				),
			},
			// Test importing the resource
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					hvnID := rs.Primary.Attributes["hvn_id"]
					privateLinkID := rs.Primary.Attributes["private_link_id"]

					return fmt.Sprintf("%s:%s", hvnID, privateLinkID), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			// Update consumer regions
			{
				Config: testAccResourcePrivateLinkUpdatedConsumerRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "private_link_id", "test-private-link"),
					resource.TestCheckResourceAttr(resourceName, "consumer_regions.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", hvnResourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_cluster_id", vaultResourceName, "cluster_id"),
				),
			},
		},
	})
}

func testAccResourcePrivateLinkBasic() string {
	return `
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
  hvn_id     = hcp_hvn.test.hvn_id
  cluster_id = "test-vault-cluster"
  tier       = "standard_small"
  public_endpoint = false
  
  major_version_upgrade_config {
    upgrade_type = "AUTOMATIC"
  }
}

resource "hcp_private_link" "test" {
  hvn_id           = hcp_hvn.test.hvn_id
  private_link_id  = "test-private-link"
  vault_cluster_id = hcp_vault_cluster.test.cluster_id

  consumer_accounts = [
    "arn:aws:iam::311485635366:root"
  ]
  
  consumer_regions = [
    "us-west-1"
  ]
}
`
}

func testAccResourcePrivateLinkUpdatedConsumerRegions() string {
	return `
resource "hcp_hvn" "test" {
  hvn_id         = "test-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
  hvn_id     = hcp_hvn.test.hvn_id
  cluster_id = "test-vault-cluster"
  tier       = "standard_small"
  public_endpoint = false
  
  major_version_upgrade_config {
    upgrade_type = "AUTOMATIC"
  }
}

resource "hcp_private_link" "test" {
  hvn_id           = hcp_hvn.test.hvn_id
  private_link_id  = "test-private-link"
  vault_cluster_id = hcp_vault_cluster.test.cluster_id

  consumer_accounts = [
    "arn:aws:iam::311485635366:root"
  ]
  
  consumer_regions = [
	"us-west-1",
    "us-east-1"
  ]
}
`
}

func testAccCheckPrivateLinkExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*clients.Client)
		privateLinkID := rs.Primary.Attributes["private_link_id"]
		hvnID := rs.Primary.Attributes["hvn_id"]
		projectID, err := GetProjectID(rs.Primary.Attributes["project_id"], client.Config.ProjectID)
		if err != nil {
			return fmt.Errorf("unable to retrieve project ID: %v", err)
		}

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      projectID,
		}

		privateLinkService, err := clients.GetPrivateLinkServiceByID(context.Background(), client, privateLinkID, hvnID, loc)
		if err != nil {
			return fmt.Errorf("error fetching private link with resource ID: %s, err: %s", rs.Primary.ID, err)
		}

		if privateLinkService.ID != privateLinkID {
			return fmt.Errorf("found private link %s, expected %s", privateLinkService.ID, privateLinkID)
		}

		return nil
	}
}

func testAccCheckPrivateLinkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcp_private_link" {
			continue
		}

		privateLinkID := rs.Primary.Attributes["private_link_id"]
		hvnID := rs.Primary.Attributes["hvn_id"]
		projectID, err := GetProjectID(rs.Primary.Attributes["project_id"], client.Config.ProjectID)
		if err != nil {
			return fmt.Errorf("unable to retrieve project ID: %v", err)
		}

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      projectID,
		}

		_, err = clients.GetPrivateLinkServiceByID(context.Background(), client, privateLinkID, hvnID, loc)
		if err == nil {
			return fmt.Errorf("private link still exists")
		}
	}

	return nil
}
