// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccGroupIamPolicyResource(t *testing.T) {
	resourceName := acctest.RandString(16)
	fullResourceName := fmt.Sprintf("iam/organization/%s/group/%s", uuid.NewString(), acctest.RandString(16))
	roleName := "roles/iam.group-manager"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupIamPolicy(resourceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_group_iam_policy.example", "policy_data", "data.hcp_iam_group_policy.example", "policy_data"),
				),
			},
			{
				ResourceName:                         "hcp_group_iam_policy.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccGroupImportID,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccGroupIamPolicy(fullResourceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_group_iam_policy.example", "policy_data", "data.hcp_iam_group_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccGroupIamBindingResource(t *testing.T) {
	resourceName := acctest.RandString(16)
	fullResourceName := fmt.Sprintf("iam/organization/%s/group/%s", uuid.NewString(), acctest.RandString(16))
	roleName := "roles/iam.group-manager"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupIamBinding(resourceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "role"),
				),
			},
			{
				Config: testAccGroupIamBinding(fullResourceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccGroupIamPolicy(resourceName, roleName string) string {
	return fmt.Sprintf(`
data "hcp_group" "example" {
	name = %q
}

data "hcp_user_principal" "example" {
	email = "test@example.com"
}

data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = %q
      principals = [
				data.hcp_user_principal.example.user_id,
      ]
    },
  ]
}

resource "hcp_group_iam_policy" "example" {
	resource_name = data.hcp_group.example.resource_name
	policy_data = data.hcp_iam_policy.example.policy_data
}
`, resourceName, roleName)
}

func testAccGroupIamBinding(resourceName, roleName string) string {
	return fmt.Sprintf(`
data "hcp_group" "example" {
	name = %q
}

data "hcp_user_principal" "example" {
	email = "test@example.com"
}

resource "hcp_group_iam_binding" "example" {
	resource_name = hcp_group.example.resource_name
	principal_id = hcp_user_principal.example.resource_id
	role = %q
}
`, resourceName, roleName)
}
