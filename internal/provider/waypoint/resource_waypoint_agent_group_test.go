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

func TestAcc_Waypoint_Agent_Group_basic(t *testing.T) {
	t.Parallel()

	var agentGroupModel waypoint.AgentGroupResourceModel
	resourceName := "hcp_waypoint_agent_group.test"
	agentGroupName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAgentGroupDestroy(t, &agentGroupModel),
		Steps: []resource.TestStep{
			{
				Config: testAgentGroup(agentGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAgentGroupExists(t, resourceName, &agentGroupModel),
					testAccCheckWaypointAgentGroupName(t, &agentGroupModel, agentGroupName),
					resource.TestCheckResourceAttr(resourceName, "name", agentGroupName),
				),
			},
		},
	})
}

func testAccCheckWaypointAgentGroupExists(t *testing.T, resourceName string, agentGroupModel *waypoint.AgentGroupResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)

		// Get the project ID, org ID, and Name from state
		projectID := rs.Primary.Attributes["project_id"]
		groupName := rs.Primary.Attributes["name"]
		orgID := rs.Primary.Attributes["organization_id"]

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			ProjectID:      projectID,
			OrganizationID: orgID,
		}

		// Retrieve the agent group using the client
		agentGroup, err := clients.GetAgentGroup(context.Background(), client, loc, groupName)
		if err != nil {
			return fmt.Errorf("error retrieving agent group %q: %w", groupName, err)
		}

		// Verify the agent group exists
		if agentGroup == nil {
			return fmt.Errorf("agent group %q not found", groupName)
		}

		agentGroupModel.Name = types.StringValue(agentGroup.Name)
		agentGroupModel.Description = types.StringValue(agentGroup.Description)

		return nil
	}
}

func testAccCheckWaypointAgentGroupName(_ *testing.T, agentGroupModel *waypoint.AgentGroupResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if agentGroupModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected agent group name to be %q, but got %q", nameValue, agentGroupModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointAgentGroupDestroy(t *testing.T, agentGroupModel *waypoint.AgentGroupResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acctest.HCPClients(t)
		name := agentGroupModel.Name.ValueString()
		projectID := agentGroupModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		agentGroupGetResponse, err := clients.GetAgentGroup(context.Background(), client, loc, name)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if agentGroupGetResponse != nil {
			return fmt.Errorf("expected agent group to be destroyed, but it still exists")
		}

		return fmt.Errorf("both agent group and error were nil in destroy check, this should not happen")
	}
}

func testAgentGroup(groupName string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_agent_group" "test" {
	name        = %q
	description = "Test Agent Group"
}
`, groupName)
}
