// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccProjectDataSource(t *testing.T) {
	project := acctest.RandString(16)
	description := acctest.RandString(64)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckProjectConfig(project, description, "resource_id"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckDataSourceStateMatchesResourceStateWithIgnores(
						"data.hcp_project.project",
						"hcp_project.project",
						map[string]struct{}{
							"project": {},
						},
					),
				),
			},
			{
				Config: testAccCheckProjectConfig(project, description, "resource_name"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckDataSourceStateMatchesResourceStateWithIgnores(
						"data.hcp_project.project",
						"hcp_project.project",
						map[string]struct{}{
							"project": {},
						},
					),
				),
			},
		},
	})
}

func testAccCheckProjectConfig(name, description, refAttr string) string {
	return fmt.Sprintf(`
resource "hcp_project" "project" {
  name        = %q
  description = %q
}

data "hcp_project" "project" {
  project = hcp_project.project.%s
}
`, name, description, refAttr)
}
