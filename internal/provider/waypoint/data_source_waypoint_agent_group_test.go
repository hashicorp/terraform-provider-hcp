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

func TestAcc_Waypoint_Agent_Group_DataSource_basic(t *testing.T) {
	t.Parallel()
	var agentGroupModel waypoint.AgentGroupResourceModel
	resourceName := "hcp_waypoint_agent_group.test"
	dataSourceName := "data." + resourceName
	agentGroupName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAgentGroupDestroy(t, &agentGroupModel),
		Steps: []resource.TestStep{
			{
				// establish the base agent group
				Config: testAgentGroup(agentGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAgentGroupExists(t, resourceName, &agentGroupModel),
				),
			},
			{
				// add a data source config to read the agent group
				Config: testDataAgentGroup(agentGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", agentGroupName),
				),
			},
		},
	})
}

func testDataAgentGroup(agentGroupName string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_agent_group" "test" {
  name    = hcp_waypoint_agent_group.test.name
}`, testAgentGroup(agentGroupName))
}
