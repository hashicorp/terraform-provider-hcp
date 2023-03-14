// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	hvnUniqueID = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))
)

var testAccAwsHvnConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}

data "hcp_hvn" "test" {
	hvn_id = hcp_hvn.test.hvn_id
}
`, hvnUniqueID)

// Currently in public beta
var testAccAzureHvnConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "azure"
	region         = "eastus"
}

data "hcp_hvn" "test" {
	hvn_id = hcp_hvn.test.hvn_id
}
`, hvnUniqueID)

// This includes tests against both the resource and the corresponding datasource
// to shorten testing time.
func TestAccAwsHvnOnly(t *testing.T) {
	resourceName := "hcp_hvn.test"
	dataSourceName := "data.hcp_hvn.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckHvnDestroy,
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testConfig(testAccAwsHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnUniqueID),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnUniqueID, HvnResourceType, resourceName),
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

					return rs.Primary.Attributes["hvn_id"], nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(testAccAwsHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnUniqueID),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnUniqueID, HvnResourceType, resourceName),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccAwsHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", dataSourceName, "cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_account_id", dataSourceName, "provider_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "self_link", dataSourceName, "self_link"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
				),
			},
		},
	})
}

func TestAccAzureHvnOnly(t *testing.T) {
	resourceName := "hcp_hvn.test"
	dataSourceName := "data.hcp_hvn.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckHvnDestroy,
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testConfig(testAccAzureHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnUniqueID),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "azure"),
					resource.TestCheckResourceAttr(resourceName, "region", "eastus"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnUniqueID, HvnResourceType, resourceName),
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

					return rs.Primary.Attributes["hvn_id"], nil
				},
				ImportStateVerify: true,
			},
			// Tests read
			{
				Config: testConfig(testAccAzureHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", hvnUniqueID),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "azure"),
					resource.TestCheckResourceAttr(resourceName, "region", "eastus"),
					resource.TestCheckResourceAttrSet(resourceName, "cidr_block"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					testLink(resourceName, "self_link", hvnUniqueID, HvnResourceType, resourceName),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccAzureHvnConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "cidr_block", dataSourceName, "cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "self_link", dataSourceName, "self_link"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
				),
			},
		},
	})
}

func testAccCheckHvnExists(name string) resource.TestCheckFunc {
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

		link, err := buildLinkFromURL(id, HvnResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		hvnID := link.ID
		loc := link.Location

		if _, err := clients.GetHvnByID(context.Background(), client, loc, hvnID); err != nil {
			return fmt.Errorf("unable to read HVN %q: %v", id, err)
		}

		return nil
	}
}

func testLink(resourceName, fieldName, expectedID, expectedType, projectIDSourceResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		selfLink, ok := rs.Primary.Attributes[fieldName]
		if !ok {
			return fmt.Errorf("%s isn't set", fieldName)
		}

		projectIDSource, ok := s.RootModule().Resources[projectIDSourceResource]
		if !ok {
			return fmt.Errorf("not found: %s", projectIDSourceResource)
		}

		projectID, ok := projectIDSource.Primary.Attributes["project_id"]
		if !ok {
			return fmt.Errorf("project_id isn't set")
		}

		link, err := linkURL(&sharedmodels.HashicorpCloudLocationLink{
			ID: expectedID,
			Location: &sharedmodels.HashicorpCloudLocationLocation{
				ProjectID: projectID},
			Type: expectedType,
		})
		if err != nil {
			return fmt.Errorf("unable to build link: %v", err)
		}

		if link != selfLink {
			return fmt.Errorf("incorrect self_link, expected: %s, got: %s", link, selfLink)
		}

		return nil
	}
}

func testAccCheckHvnDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_hvn":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, HvnResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			hvnID := link.ID
			loc := link.Location

			_, err = clients.GetHvnByID(context.Background(), client, loc, hvnID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
