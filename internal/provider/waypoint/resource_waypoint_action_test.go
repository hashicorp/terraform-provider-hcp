// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAccWaypoint_Action_basic(t *testing.T) {
	// Skip this test unless the appropriate environment variable is set
	// This is to prevent running this test by default
	if os.Getenv("HCP_WAYP_ACTION_TEST") == "" {
		t.Skipf("Waypoint Action tests skipped unless env '%s' set",
			"HCP_WAYP_ACTION_TEST")
		return
	}
	var actionCfgModel waypoint.ActionResourceModel
	resourceName := "hcp_waypoint_action.test"
	actionName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointActionDestroy(t, &actionCfgModel),
		Steps: []resource.TestStep{
			{
				Config: testAction(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointActionExists(t, resourceName, &actionCfgModel),
					testAccCheckWaypointActionName(t, &actionCfgModel, actionName),
					resource.TestCheckResourceAttr(resourceName, "name", actionName),
				),
			},
		},
	})
}

// Simple attribute check on the action received from the API
func testAccCheckWaypointActionName(_ *testing.T, actionCfgModel *waypoint.ActionResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if actionCfgModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected action name to be %s, but got %s", nameValue, actionCfgModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointActionExists(t *testing.T, resourceName string, actionCfgModel *waypoint.ActionResourceModel) resource.TestCheckFunc {
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
		actionName := rs.Primary.Attributes["name"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the action config
		actionCfg, err := clients.GetAction(context.Background(), client, loc, actionID, actionName)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name and id
		if actionCfgModel != nil {
			actionCfgModel.Name = types.StringValue(actionCfg.Name)
			actionCfgModel.ID = types.StringValue(actionCfg.ID)
		}

		return nil
	}
}

func testAccCheckWaypointActionDestroy(t *testing.T, actionModel *waypoint.ActionResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := actionModel.ID.ValueString()
		name := actionModel.Name.ValueString()
		projectID := actionModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		actionGetResp, err := clients.GetAction(context.Background(), client, loc, id, name)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if actionGetResp != nil {
			return fmt.Errorf("expected action to be destroyed, but it still exists")
		}

		return fmt.Errorf("both action and error were nil in destroy check, this should not happen")
	}
}

func testAction(actionName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_action" "test" {
	name = "%s"
	description = "Test action"
	request = {
	    custom = {
			method = "GET"
			url = "https://example.com"
			headers = {
				Test-Header = "test"
			}
			body = "test"
		}
	}
}
`, actionName)
}
