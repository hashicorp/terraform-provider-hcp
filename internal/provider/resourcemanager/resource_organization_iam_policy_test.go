// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccOrganizationIamBindingResource(t *testing.T) {
	t.Parallel()

	spName := acctest.RandString(16)
	roleName := "roles/contributor"
	roleName2 := "roles/viewer"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationIamBinding(spName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_organization_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_organization_iam_binding.example", "role"),
				),
			},
			{
				Config: testAccOrganizationIamBinding(spName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_organization_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_organization_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccOrganizationIamBinding(spName, roleName string) string {
	return fmt.Sprintf(`
data "hcp_organization" "example" { }

resource "hcp_service_principal" "example" {
	name = %q
	parent = data.hcp_organization.example.resource_name
}

resource "hcp_organization_iam_binding" "example" {
	principal_id = hcp_service_principal.example.resource_id
	role = %q
}
`, spName, roleName)
}
