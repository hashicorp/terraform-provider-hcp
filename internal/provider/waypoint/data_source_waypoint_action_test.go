// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAcc_Waypoint_Action_DataSource_basic(t *testing.T) {
	t.Parallel()

	// Skip this test unless the appropriate environment variable is set
	// This is to prevent running this test by default
	if os.Getenv("HCP_WAYP_ACTION_TEST") == "" {
		t.Skipf("Waypoint Action tests skipped unless env '%s' set",
			"HCP_WAYP_ACTION_TEST")
		return
	}
	var actionModel waypoint.ActionResourceModel
	resourceName := "hcp_waypoint_action.test"
	dataSourceName := "data." + resourceName
	actionName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointActionDestroy(t, &actionModel),
		Steps: []resource.TestStep{
			{
				// establish the base action config
				Config: testAction(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointActionExists(t, resourceName, &actionModel),
				),
			},
			{
				// add a data source config to read the action config
				Config: testDataActionConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", actionName),
				),
			},
		},
	})
}

func testDataActionConfig(actionName string) string {
	return fmt.Sprintf(`%s

data "hcp_waypoint_action" "test" {
  name    = hcp_waypoint_action.test.name
}`, testAction(actionName))
}
