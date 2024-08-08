// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"context"
	"fmt"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAccWaypoint_Application_basic(t *testing.T) {
	var applicationModel waypoint.ApplicationResourceModel
	resourceName := "hcp_waypoint_application.test"
	templateName := generateRandomName()
	applicationName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				Config: testApplicationConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointApplicationName(t, &applicationModel, applicationName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationName),
				),
			},
		},
	})
}

func TestAccWaypoint_ApplicationInputVariables(t *testing.T) {
	var applicationModel waypoint.ApplicationResourceModel
	resourceName := "hcp_waypoint_application.test_var_opts"
	templateName := generateRandomName()
	applicationName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				Config: testApplicationWithInputVarsConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointApplicationName(t, &applicationModel, applicationName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationName),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.0.name", "faction"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.0.value", "brotherhood-of-steel"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.0.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.1.name", "vault_dweller_name"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.1.value", "courier"),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.1.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.0.name", "waypoint_application"),
				),
			},
		},
	})
}

func TestAccWaypoint_ApplicationInputVariables_OnTemplate(t *testing.T) {
	var applicationModel waypoint.ApplicationResourceModel
	resourceName := "hcp_waypoint_application.test_var_opts"
	templateName := generateRandomName()
	applicationName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				Config: testApplicationWithNoInputVarsConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointApplicationName(t, &applicationModel, applicationName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationName),
					resource.TestCheckResourceAttr(resourceName, "application_input_variables.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.0.name", "faction"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.0.value", "brotherhood-of-steel"),
					// resource.TestCheckResourceAttr(resourceName, "template_input_variables.0.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.1.name", "vault_dweller_name"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.1.value", "lone-wanderer"),
					// resource.TestCheckResourceAttr(resourceName, "template_input_variables.1.variable_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.2.name", "waypoint_application"),
					resource.TestCheckResourceAttr(resourceName, "template_input_variables.2.value", applicationName),
				),
			},
		},
	})
}

// simple attribute check on the application receved from the API
func testAccCheckWaypointApplicationName(_ *testing.T, applicationModel *waypoint.ApplicationResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if applicationModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected application name to be %s, but got %s", nameValue, applicationModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointApplicationExists(t *testing.T, resourceName string, applicationModel *waypoint.ApplicationResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID and ID from state
		projectID := rs.Primary.Attributes["project_id"]
		appID := rs.Primary.Attributes["id"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the application
		application, err := clients.GetApplicationByID(context.Background(), client, loc, appID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name,id, and project id for now
		if applicationModel != nil {
			applicationModel.Name = types.StringValue(application.Name)
			applicationModel.ID = types.StringValue(application.ID)
			applicationModel.ProjectID = types.StringValue(projectID)
		}

		return nil
	}
}

func testAccCheckWaypointApplicationDestroy(t *testing.T, applicationModel *waypoint.ApplicationResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := applicationModel.ID.ValueString()
		projectID := applicationModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		application, err := clients.GetApplicationByID(context.Background(), client, loc, id)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if application != nil {
			return fmt.Errorf("expected application to be destroyed, but it still exists")
		}

		return fmt.Errorf("both application and error were nil in destroy check, this should not happen")
	}
}

// these are hardcoded project and no-code module values because they work. The
// automated tests do not run acceptance tests at this time, so these should be
// sufficient for now.
func testApplicationConfig(tempName, appName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test" {
  name    = "%s"
  summary = "some summary for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module = {
    source  = "private/waypoint-tfc-testing/waypoint-template-starter/null"
    version = "0.0.2"
  }
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["one", "two"]
}

resource "hcp_waypoint_application" "test" {
  name    = "%s"
  template_id = hcp_waypoint_template.test.id
}`, tempName, appName)
}

func testApplicationWithInputVarsConfig(tempName, appName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test_var_opts" {
  name    = "%s"
  summary = "some summary for fun"
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

resource "hcp_waypoint_application" "test_var_opts" {
  name    = "%s"
  template_id = hcp_waypoint_template.test_var_opts.id

  application_input_variables = [
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
}`, tempName, appName)
}

func testApplicationWithNoInputVarsConfig(tempName, appName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test_var_opts" {
  name    = "%s"
  summary = "some summary for fun"
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

resource "hcp_waypoint_application" "test_var_opts" {
  name    			      = "%s"
  template_id = hcp_waypoint_template.test_var_opts.id
}`, tempName, appName)
}
