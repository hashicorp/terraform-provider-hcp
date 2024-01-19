// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccProjectIamPolicyResource(t *testing.T) {
	projectName := acctest.RandString(16)
	roleName := "roles/contributor"
	roleName2 := "roles/viewer"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectIamPolicy(projectName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "project_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_project_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
			{
				ResourceName:                         "hcp_project_iam_policy.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "project_id",
				ImportStateIdFunc:                    testAccProjectImportID,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccProjectIamPolicy(projectName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "project_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_project_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccProjectIamBindingResource(t *testing.T) {
	projectName := acctest.RandString(16)
	roleName := "roles/contributor"
	roleName2 := "roles/viewer"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectIamBinding(projectName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "project_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "role"),
				),
			},
			{
				Config: testAccProjectIamBinding(projectName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "project_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_project_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccProjectIamPolicy(projectName, roleName string) string {
	return fmt.Sprintf(`
resource "hcp_project" "example" {
	name = %q
}

resource "hcp_service_principal" "example" {
	name = "test-sp"
	parent = hcp_project.example.resource_name
}

data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = %q
      principals = [
		hcp_service_principal.example.resource_id,
      ]
    },
  ]
}

resource "hcp_project_iam_policy" "example" {
	project_id = hcp_project.example.resource_id
	policy_data = data.hcp_iam_policy.example.policy_data
}
`, projectName, roleName)
}

func testAccProjectIamBinding(projectName, roleName string) string {
	return fmt.Sprintf(`
resource "hcp_project" "example" {
	name = %q
}

resource "hcp_service_principal" "example" {
	name = "test-sp"
	parent = hcp_project.example.resource_name
}

resource "hcp_project_iam_binding" "example" {
	project_id = hcp_project.example.resource_id
	principal_id = hcp_service_principal.example.resource_id
	role = %q
}
`, projectName, roleName)
}
