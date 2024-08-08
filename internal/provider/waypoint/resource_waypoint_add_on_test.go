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

func TestAccWaypoint_Add_On_basic(t *testing.T) {
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test"
	addOnName := generateRandomName()
	templateName := generateRandomName()
	appName := generateRandomName()
	defName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
					testAccCheckWaypointAddOnName(t, &addOnModel, addOnName),
					resource.TestCheckResourceAttr(resourceName, "name", addOnName),
				),
			},
		},
	})
}

func TestAccWaypoint_AddOnInputVariables(t *testing.T) {
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test_var_opts"
	addOnName := generateRandomName()
	templateName := generateRandomName()
	appName := generateRandomName()
	defName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnWithInputVarsConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
					testAccCheckWaypointAddOnName(t, &addOnModel, addOnName),
					resource.TestCheckResourceAttr(resourceName, "name", addOnName),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.0.name", "faction"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.0.value", "brotherhood-of-steel"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.0.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.1.name", "vault_dweller_name"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.1.value", "courier"),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.1.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "add_on_definition_input_variables.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "add_on_definition_input_variables.0.name", "waypoint_add_on"),
					resource.TestCheckResourceAttr(resourceName, "add_on_definition_input_variables.1.name", "waypoint_add_on_definition"),
					resource.TestCheckResourceAttr(resourceName, "add_on_definition_input_variables.2.name", "waypoint_application"),
				),
			},
		},
	})
}

func TestAccWaypoint_AddOnInputVariables_OnDefinition(t *testing.T) {
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test_var_opts"
	addOnName := generateRandomName()
	templateName := generateRandomName()
	appName := generateRandomName()
	defName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnWithNoInputVarsConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
					testAccCheckWaypointAddOnName(t, &addOnModel, addOnName),
					resource.TestCheckResourceAttr(resourceName, "name", addOnName),
					resource.TestCheckResourceAttr(resourceName, "add_on_input_variables.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "add_on_definition_input_variables.#", "5"),
				),
			},
		},
	})
}

// simple attribute check on the add-on definition received from the API
func testAccCheckWaypointAddOnName(t *testing.T, addOnModel *waypoint.AddOnResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if addOnModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected add-on name to be %s, but got %s", nameValue, addOnModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointAddOnExists(t *testing.T, resourceName string, addOnModel *waypoint.AddOnResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID and ID from state
		projectID := rs.Primary.Attributes["project_id"]
		addOnID := rs.Primary.Attributes["id"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the add-on
		addOn, err := clients.GetAddOnByID(context.Background(), client, loc, addOnID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name and ID for now
		addOnModel.Name = types.StringValue(addOn.Name)
		addOnModel.ID = types.StringValue(addOn.ID)

		return nil
	}
}

func testAccCheckWaypointAddOnDestroy(t *testing.T, addOnModel *waypoint.AddOnResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := addOnModel.ID.ValueString()
		projectID := client.Config.ProjectID
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		addOn, err := clients.GetAddOnByID(context.Background(), client, loc, id)
		if err != nil {
			// expected (500 because the application is destroyed as well)
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if addOn != nil {
			return fmt.Errorf("expected add-on to be destroyed, but it still exists")
		}

		return fmt.Errorf("both add-on and error were nil in destroy check, this should not happen")
	}
}

// These are hardcoded project and no-code module values because they work. The
// automated tests do not run acceptance tests at this time, so these should be
// sufficient for now.
func testAddOnConfig(templateName string, appName string, defName string, addOnName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test" {
  name    = "%s"
  summary = "some summary for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["one", "two"]
}

resource "hcp_waypoint_application" "test" {
  name    = "%s"
  template_id = hcp_waypoint_template.test.id
}

resource "hcp_waypoint_add_on_definition" "test" {
  name    = "%s"
  summary = "some summary for fun"
  description = "some description for fun"
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
}

resource "hcp_waypoint_add_on" "test" {
  name    = "%s"
  application_id = hcp_waypoint_application.test.id
  definition_id = hcp_waypoint_add_on_definition.test.id
}`, templateName, appName, defName, addOnName)
}

func testAddOnWithInputVarsConfig(tempName, appName, defName, addOnName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test" {
  name    = "%s"
  summary = "some summary for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["one", "two"]
}

resource "hcp_waypoint_application" "test" {
  name    = "%s"
  template_id = hcp_waypoint_template.test.id
}

resource "hcp_waypoint_add_on_definition" "test_var_opts" {
  name    = "%s"
  summary = "some summary for fun"
  description = "some description for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-vault-dweller/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["fallout", "vault-tec"]
  variable_options = [
	{
	  name          = "vault_dweller_name"
      variable_type = "string"
      user_editable = true
      options 		= [
        "lucy",
        "courier",
        "lone-wanderer",
        "sole-survivor",
      ]
    },
    {
	  name          = "faction"
      variable_type = "string"
      user_editable = true
      options 		= [
        "ncr",
        "brotherhood-of-steel",
        "caesars-legion",
        "raiders",
        "institute"
      ]
    },
  ]
}

resource "hcp_waypoint_add_on" "test_var_opts" {
  name           = "%s"
  definition_id  = hcp_waypoint_add_on_definition.test_var_opts.id
  application_id = hcp_waypoint_application.test.id

  add_on_input_variables = [
	{
      name  		= "faction"
      variable_type = "string"
      value 		= "brotherhood-of-steel"
    },
    {
      name  		= "vault_dweller_name"
      variable_type = "string"
	  value 		= "courier"
    }	
  ]
}

`, tempName, appName, defName, addOnName)
}

func testAddOnWithNoInputVarsConfig(tempName, appName, defName, addOnName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test" {
  name    = "%s"
  summary = "some summary for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["one", "two"]
}

resource "hcp_waypoint_application" "test" {
  name    = "%s"
  template_id = hcp_waypoint_template.test.id
}

resource "hcp_waypoint_add_on_definition" "test_var_opts" {
  name        = "%s"
  summary     = "some summary for fun"
  description = "some description"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module = "private/waypoint-tfc-testing/waypoint-vault-dweller/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["fallout", "vault-tec"]
  variable_options = [
	{
	  name          = "vault_dweller_name"
      variable_type = "string"
      user_editable = false
      options 		= [
        "lone-wanderer",
      ]
    },
    {
	  name          = "faction"
      variable_type = "string"
      user_editable = false
      options 		= [
        "brotherhood-of-steel",
      ]
    },
  ]
}

resource "hcp_waypoint_add_on" "test_var_opts" {
  name    		 = "%s"
  definition_id  = hcp_waypoint_add_on_definition.test_var_opts.id
  application_id = hcp_waypoint_application.test.id
}`, tempName, appName, defName, addOnName)
}
