// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccGroupResource(t *testing.T) {
	groupName := acctest.RandString(16)
	description := acctest.RandString(8)

	groupNameUpdated := acctest.RandString(16)
	descriptionUpdated := acctest.RandString(200)
	var group models.HashicorpCloudIamGroup

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckGroupDestroy(t, groupName),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfigResourceWithoutDescription(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group.example", "display_name", groupName),
					resource.TestCheckResourceAttr("hcp_group.example", "description", ""),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_id"),
					testAccGroupExists(t, "hcp_group.example", &group),
				),
			},
			{
				Config: testAccGroupConfigResource(groupName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group.example", "display_name", groupName),
					resource.TestCheckResourceAttr("hcp_group.example", "description", description),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_id"),
					testAccGroupExists(t, "hcp_group.example", &group),
				),
			},
			{
				Config: testAccGroupConfigResourceWithoutDescription(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group.example", "display_name", groupName),
					resource.TestCheckResourceAttr("hcp_group.example", "description", ""),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_id"),
					testAccGroupExists(t, "hcp_group.example", &group),
				),
			},
			{
				ResourceName:                         "hcp_group.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccGroupImportID,
				ImportStateVerify:                    true,
			},
			{
				// Update the name/description
				Config: testAccGroupConfigResource(groupNameUpdated, descriptionUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_group.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group.example", "display_name", groupNameUpdated),
					resource.TestCheckResourceAttr("hcp_group.example", "description", descriptionUpdated),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group.example", "resource_id"),
				),
			},
		},
	})
}

func testAccGroupConfigResource(displayName, description string) string {
	return fmt.Sprintf(`
	resource "hcp_group" "example" {
		display_name = %q
		description = %q
	}
`, displayName, description)
}

func testAccGroupConfigResourceWithoutDescription(displayName string) string {
	return fmt.Sprintf(`
	resource "hcp_group" "example" {
		display_name = %q
	}
`, displayName)
}

// testAccGroupsExists queries the API and retrieves the matching
// group.
func testAccGroupExists(t *testing.T, resourceName string, sp *models.HashicorpCloudIamGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Get the group resource name from state
		rname := rs.Primary.Attributes["resource_name"]

		// Fetch the group
		client := acctest.HCPClients(t)
		getParams := groups_service.NewGroupsServiceGetGroupParams()
		getParams.ResourceName = rname
		res, err := client.Groups.GroupsServiceGetGroup(getParams, nil)
		if err != nil {
			return err
		}

		if res.GetPayload().Group == nil {
			return fmt.Errorf("Group (%s) not found", rname)
		}

		// assign the response project to the pointer
		*sp = *res.GetPayload().Group
		return nil
	}
}

// testAccGroupImportID retrieves the resource_name so that it can be imported.
func testAccGroupImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_group.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_name"]
	if !ok {
		return "", fmt.Errorf("resource_name not set")
	}

	return id, nil
}

func testAccCheckGroupDestroy(t *testing.T, groupName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		resourceName := fmt.Sprintf("iam/organization/%s/group/%s", client.GetOrganizationID(), groupName)
		getParams := groups_service.NewGroupsServiceGetGroupParams().WithResourceName(resourceName)
		_, err := client.Groups.GroupsServiceGetGroup(getParams, nil)
		if err != nil {
			var getErr *groups_service.GroupsServiceGetGroupDefault
			if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {

				return nil
			}

		}
		return fmt.Errorf("didn't get a 404 when reading destroyed group %s: %v", resourceName, err)
	}
}
