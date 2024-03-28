// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccGroupDataSource(t *testing.T) {
	dataSourceAddress := "data.hcp_group.test"
	groupName := acctest.RandString(16)
	description := acctest.RandString(64)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckGroupDestroy(t, groupName),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(groupName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "resource_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "description"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "display_name"),
				),
			},
		},
	})
}

func testAccGroupConfig(name, description string) string {
	return fmt.Sprintf(`
	resource "hcp_group" "test" { 
		display_name = %q
		description = %q
	}
	data "hcp_group" "test" { 
		resource_name = hcp_group.test.resource_name
	}
`, name, description)
}
