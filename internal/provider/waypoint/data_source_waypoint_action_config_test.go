// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAccWaypoint_Action_Config_DataSource_basic(t *testing.T) {
	var actionConfigModel waypoint.ActionConfigResourceModel
	resourceName := "hcp_waypoint_action_config.test"
	dataSourceName := "data." + resourceName
	actionName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointActionConfigDestroy(t, &actionConfigModel),
		Steps: []resource.TestStep{
			{
				// establish the base action config
				Config: testActionConfigConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointActionExists(t, resourceName, &actionConfigModel),
				),
			},
			{
				// add a data source config to read the action config
				Config: testDataActionConfigConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", actionName),
				),
			},
		},
	})
}

func testDataActionConfigConfig(actionName string) string {
	return fmt.Sprintf(`%s

data "hcp_waypoint_action_config" "test" {
  name    = hcp_waypoint_action_config.test.name
}`, testActionConfigConfig(action))
}
