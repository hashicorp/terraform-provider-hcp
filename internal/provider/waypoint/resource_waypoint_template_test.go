// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

func TestAccWaypoint_Template_basic(t *testing.T) {
	var appTemplateModel waypoint.TemplateResourceModel
	resourceName := "hcp_waypoint_template.test"
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointTemplateDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				Config: testTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointTemplateName(t, &appTemplateModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				Config: testTemplateConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointTemplateName(t, &appTemplateModel, updatedName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
				),
			},
		},
	})
}

func TestAccWaypoint_template_with_variable_options(t *testing.T) {
	var appTemplateModel waypoint.TemplateResourceModel
	resourceName := "hcp_waypoint_template.var_opts_test"
	name := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointTemplateDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				Config: testTemplateConfigWithVarOpts(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointTemplateName(t, &appTemplateModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
		},
	})
}

// simple attribute check on the template receved from the API
func testAccCheckWaypointTemplateName(t *testing.T, appTemplateModel *waypoint.TemplateResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if appTemplateModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected template name to be %s, but got %s", nameValue, appTemplateModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointTemplateExists(t *testing.T, resourceName string, appTemplateModel *waypoint.TemplateResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
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

		// Fetch the template
		template, err := clients.GetApplicationTemplateByID(context.Background(), client, loc, appTempID)
		if err != nil {
			return err
		}

		// at this time we're only verifying existence and not checking all the
		// values, so only set name,id, and project id for now
		if appTemplateModel != nil {
			appTemplateModel.Name = types.StringValue(template.Name)
			appTemplateModel.ID = types.StringValue(template.ID)
			appTemplateModel.ProjectID = types.StringValue(projectID)
		}

		return nil
	}
}

func testAccCheckWaypointTemplateDestroy(t *testing.T, appTemplateModel *waypoint.TemplateResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := appTemplateModel.ID.ValueString()
		projectID := appTemplateModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		template, err := clients.GetApplicationTemplateByID(context.Background(), client, loc, id)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if template != nil {
			return fmt.Errorf("expected Template to be destroyed, but it still exists")
		}

		return fmt.Errorf("both template and error were nil in destroy check, this should not happen")
	}
}

func testTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "test" {
  name                     = "%s"
  summary                  = "some summary for fun"
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-template-starter/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  labels = ["one", "two"]
}`, name)
}

func testTemplateConfigWithVarOpts(name string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_template" "var_opts_test" {
  name                     = "%s"
  summary                  = "A template with a variable with options."
  readme_markdown_template = base64encode("# Some Readme")
  terraform_no_code_module_source = "private/waypoint-tfc-testing/waypoint-vault-dweller/null"
  terraform_cloud_workspace_details = {
    name                 = "Default Project"
    terraform_project_id = "prj-gfVyPJ2q2Aurn25o"
  }
  variable_options = [
	{
	  name          = "vault_dweller_name"
	  variable_type = "string"
      user_editable = true
      options       = []
    },
    {
      name          = "faction"
      variable_type = "string"
      user_editable = true
      options       = [
        "ncr",
        "brotherhood-of-steel",
        "caesars-legion",
        "raiders",
        "institute"
      ]
    }
  ]
}`, name)
}

// generateRandomName will create a valid randomized name
func generateRandomName() string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return "hcp-provider-acctest-" + string(b)
}
