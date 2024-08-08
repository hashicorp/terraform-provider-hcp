// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

func TestAccWaypoint_Add_On_Definition_basic(t *testing.T) {
	var addOnDefinitionModel waypoint.AddOnDefinitionResourceModel
	resourceName := "hcp_waypoint_add_on_definition.test"
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDefinitionDestroy(t, &addOnDefinitionModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnDefinitionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnDefinitionExists(t, resourceName, &addOnDefinitionModel),
					testAccCheckWaypointAddOnDefinitionName(t, &addOnDefinitionModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.name", "string_variable"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.options.0", "b"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.user_editable", "false"),
				),
			},
			{
				Config: testAddOnDefinitionConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnDefinitionExists(t, resourceName, &addOnDefinitionModel),
					testAccCheckWaypointAddOnDefinitionName(t, &addOnDefinitionModel, updatedName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.name", "string_variable"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.options.0", "b"),
					resource.TestCheckResourceAttr(resourceName, "variable_options.0.user_editable", "false"),
				),
			},
		},
	})
}

// simple attribute check on the add-on definition received from the API
func testAccCheckWaypointAddOnDefinitionName(t *testing.T, addOnDefinitionModel *waypoint.AddOnDefinitionResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if addOnDefinitionModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected add-on definition name to be %s, but got %s", nameValue, addOnDefinitionModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointAddOnDefinitionExists(t *testing.T, resourceName string, definitionModel *waypoint.AddOnDefinitionResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID and ID from state
		projectID := rs.Primary.Attributes["project_id"]
		appTempID := rs.Primary.Attributes["id"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the add-on definition
		definition, err := clients.GetAddOnDefinitionByID(context.Background(), client, loc, appTempID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name,id, and project id for now
		definitionModel.Name = types.StringValue(definition.Name)
		definitionModel.ID = types.StringValue(definition.ID)
		definitionModel.ProjectID = types.StringValue(projectID)

		return nil
	}
}

func testAccCheckWaypointAddOnDefinitionDestroy(t *testing.T, definitionModel *waypoint.AddOnDefinitionResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := definitionModel.ID.ValueString()
		projectID := definitionModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		definition, err := clients.GetAddOnDefinitionByID(context.Background(), client, loc, id)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if definition != nil {
			return fmt.Errorf("expected add-on definition to be destroyed, but it still exists")
		}

		return fmt.Errorf("both add-on definition and error were nil in destroy check, this should not happen")
	}
}

// TODO: Add remaining add-on definition fields to test (tags, labels, readmemarkdown.. etc)
func testAddOnDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_add_on_definition" "test" {
  name    = %q
  summary = "some summary for fun"
  description = "some description for fun"
  terraform_no_code_module = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  variable_options = [
	{
	  name        = "string_variable"
      variable_type = "string"
      options = [
        "b"
      ]
      user_editable = false
    }
  ]
}`, name)
}
