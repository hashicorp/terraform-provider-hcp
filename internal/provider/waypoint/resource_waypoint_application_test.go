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
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				Config: testApplicationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointApplicationName(t, &applicationModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				Config: testApplicationConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointApplicationName(t, &applicationModel, updatedName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
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

func testApplicationConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "hcp_waypoint_application" "test" {
  name    = %q
  application_template_id = hcp_waypoint_application_template.test.id
}`, testAppTemplateConfig(name), name)
}
