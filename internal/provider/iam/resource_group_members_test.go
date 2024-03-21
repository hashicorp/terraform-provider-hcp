// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccGroupMembersResource(t *testing.T) {
	// Test values for our integration tests in int.
	groupName := "iam/organization/d11d7309-5072-44f9-aaea-c8f37c09a8b5/group/group_members_terraform_resource_test"
	up1 := "4a836041-72f5-442d-a52f-af9e69f5a7f0"
	up2 := "626f4cb9-e666-4318-a7a9-a3a3ccb6e5f1"
	up3 := "353d4eca-5cfa-443c-94a7-32445a6928fa"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)

			// Ensure the group is empty before running the test
			members := getGroupMembers(t, groupName)
			if len(members) > 0 {
				cleanupGroupMembers(t, groupName, members)
			}
		},
		CheckDestroy: testAccCheckGroupMembersMatch(t, groupName),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembersResourceConfig(t, groupName, up1, up2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group_members.example", "group", groupName),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.0", up1),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.1", up2),
					testAccCheckGroupMembersMatch(t, groupName, up1, up2),
				),
			},
			{
				ResourceName:                         "hcp_group_members.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "group",
				ImportStateId:                        groupName,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccGroupMembersResourceConfig(t, groupName, up1, up3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group_members.example", "group", groupName),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.0", up1),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.1", up3),
					testAccCheckGroupMembersMatch(t, groupName, up1, up3),
				),
			},
			{
				// Set group membership outside of terraform to ensure that the next step brings things in line
				Config: testAccGroupMembersResourceConfig(t, groupName, up1, up3),
				Check: func(s *terraform.State) error {
					updateParams := groups_service.NewGroupsServiceUpdateGroupMembersParams()
					updateParams.SetResourceName(groupName)
					updateParams.SetBody(groups_service.GroupsServiceUpdateGroupMembersBody{
						MemberPrincipalIdsToAdd:    []string{up2},
						MemberPrincipalIdsToRemove: []string{up3},
					})

					client := acctest.HCPClients(t)
					_, err := client.Groups.GroupsServiceUpdateGroupMembers(updateParams, nil)
					return err
				},
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupMembersResourceConfig(t, groupName, up2, up3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_group_members.example", "group", groupName),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.0", up2),
					resource.TestCheckResourceAttr("hcp_group_members.example", "members.1", up3),
					testAccCheckGroupMembersMatch(t, groupName, up2, up3),
				),
			},
		},
	})
}

func testAccGroupMembersResourceConfig(t *testing.T, groupName string, principalIDs ...string) string {
	if len(principalIDs) == 0 {
		t.Fatal("at least one principal ID must be provided")
	}

	return fmt.Sprintf(`
resource "hcp_group_members" "example" {
	group = "%s"
	members = ["%s"]
}
`, groupName, strings.Join(principalIDs, `", "`))
}

func testAccCheckGroupMembersMatch(t *testing.T, groupName string, principalIDs ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		groupMembers := getGroupMembers(t, groupName)
		if len(groupMembers) != len(principalIDs) {
			return fmt.Errorf("HCP Check: expected %d members, got %d", len(principalIDs), len(groupMembers))
		}

		for _, principalID := range principalIDs {
			found := false
			for _, member := range groupMembers {
				if member == principalID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("HCP Check: expected member %s not found", principalID)
			}
		}

		return nil
	}
}

func getGroupMembers(t *testing.T, groupName string) []string {
	client := acctest.HCPClients(t)
	listParams := groups_service.NewGroupsServiceListGroupMembersParams().WithResourceName(groupName)
	res, err := client.Groups.GroupsServiceListGroupMembers(listParams, nil)
	if err != nil {
		t.Fatal(err)
	}

	members := res.GetPayload().Members
	memberIDs := make([]string, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	return memberIDs
}

func cleanupGroupMembers(t *testing.T, groupName string, members []string) {
	client := acctest.HCPClients(t)
	updateParams := groups_service.NewGroupsServiceUpdateGroupMembersParams()
	updateParams.SetResourceName(groupName)
	updateParams.SetBody(groups_service.GroupsServiceUpdateGroupMembersBody{
		MemberPrincipalIdsToRemove: members,
	})

	_, err := client.Groups.GroupsServiceUpdateGroupMembers(updateParams, nil)
	if err != nil {
		t.Fatal(err)
	}
}
