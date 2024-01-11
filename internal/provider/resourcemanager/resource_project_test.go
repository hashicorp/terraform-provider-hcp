// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccProjectResource(t *testing.T) {
	var p models.HashicorpCloudResourcemanagerProject
	projectName := acctest.RandString(16)
	description := acctest.RandString(200)

	projectNameUpdated := acctest.RandString(16)
	descriptionUpdated := acctest.RandString(200)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProject(projectName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_project.example", "name", projectName),
					resource.TestCheckResourceAttr("hcp_project.example", "description", description),
					resource.TestCheckResourceAttrSet("hcp_project.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_project.example", "resource_id"),
					testAccProjectResourceExists(t, "hcp_project.example", &p),
					testAccCheckProjectValues(&p, projectName, description),
				),
			},
			{
				ResourceName:                         "hcp_project.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_id",
				ImportStateIdFunc:                    testAccProjectImportID,
				ImportStateVerify:                    true,
			},
			{
				// Update the name / description
				Config: testAccProject(projectNameUpdated, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_project.example", "name", projectNameUpdated),
					resource.TestCheckResourceAttr("hcp_project.example", "description", descriptionUpdated),
					resource.TestCheckResourceAttrSet("hcp_project.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_project.example", "resource_id"),
					testAccProjectResourceExists(t, "hcp_project.example", &p),
					testAccCheckProjectValues(&p, projectNameUpdated, descriptionUpdated),
				),
			},
		},
	})
}

// testAccProjectImportID retrieves the resource_id so that it can be imported.
func testAccProjectImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_project.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_id"]
	if !ok {
		return "", fmt.Errorf("resource_id not set")
	}

	return id, nil
}

func testAccProject(name, description string) string {
	return fmt.Sprintf(`
resource "hcp_project" "example" {
	name = %q
	description = %q
}`, name, description)
}

// testAccProjectResourceExists queries the API and retrieves the matching
// project.
func testAccProjectResourceExists(t *testing.T, resourceName string, project *models.HashicorpCloudResourcemanagerProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Get the project ID from state
		id := rs.Primary.Attributes["resource_id"]

		// Fetch the project
		client := acctest.HCPClients(t)
		getParams := project_service.NewProjectServiceGetParams()
		getParams.ID = id
		res, err := client.Project.ProjectServiceGet(getParams, nil)
		if err != nil {
			return err
		}

		if res.GetPayload().Project == nil {
			return fmt.Errorf("Project (%s) not found", id)
		}

		// assign the response project to the pointer
		*project = *res.GetPayload().Project
		return nil
	}
}

func testAccCheckProjectValues(project *models.HashicorpCloudResourcemanagerProject, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if project.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, project.Name)
		}

		if project.Description != description {
			return fmt.Errorf("bad description, expected \"%s\", got: %#v", description, project.Description)
		}
		return nil
	}
}
