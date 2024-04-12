// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"fmt"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAccWaypoint_Action_Config_basic(t *testing.T) {
	var actionCfgModel waypoint.ActionConfigResourceModel
	resourceName := "hcp_waypoint_action_config.test"
	actionName := generateRandomName()

	resource.Test(t,, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointActionConfigDestroy(t, &actionCfgModel),
		Steps: []resource.TestStep{
			{
				Config: testActionConfigConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointActionConfigExists(t, resourceName, &actionCfgModel),
					testAccCheckWaypointActionConfigName(t, &actionCfgModel, actionName),
					resource.TestCheckResourceAttr(resourceName, "name", actionName),
				),
			},
		},
	})
}

// Simple attribute check on the action config received from the API
func testAccCheckWaypointActionConfigName(_ *testing.T, actionCfgModel *waypoint.ActionConfigResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if actionCfgModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected action config name to be %s, but got %s", nameValue, actionCfgModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointActionConfigExists(t *testing.T, resourceName string, actionCfgModel *waypoint.ActionConfigResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID and ID from state
		projectID := rs.Primary.Attributes["project_id"]
		actionID := rs.Primary.Attributes["id"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the action config
		actionCfg, err := clients.GetActionConfig(client, loc, actionID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name,id, and project id for now
		if actionCfgModel != nil {
			actionCfgModel.Name = types.StringValue(actionCfg.Name)
			actionCfgModel.ID = types.StringValue(actionCfg.ID)
			actionCfgModel.ProjectID = types.StringValue(actionCfg.ProjectID)
		}

		return nil
	}
}

func testAccCheckWaypointActionConfigDestroy(t *testing.T, actionConfigModel *waypoint.ActionConfigResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := actionConfigModel.ID.ValueString()
		projectID := actionConfigModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		actionConfig, err := clients.GetActionConfig(client, loc id)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if actionConfig != nil {
			return fmt.Errorf("expected action config to be destroyed, but it still exists")
		}

		return fmt.Errorf("both action config and error were nil in destroy check, this should not happen")
	}
}

// TODO: Fill this out with remaining information
func testActionConfigConfig(actionName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_action_config" "test" {
	name = "%s"
}
`, actionName)
}