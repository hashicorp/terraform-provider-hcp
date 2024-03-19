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

func TestAccWaypointData_Add_On_basic(t *testing.T) {
	// this is only used to verify the add-on gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				// establish the base add-on
				// note this reuses the config method from the add-on
				// resource test
				Config: testAddOnConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
				),
			},
			{
				// add a data source config to read the add-on
				Config: testDataAddOnConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
				),
			},
			{
				// update the add-on name, make sure it reflects in the data source
				Config: testDataAddOnConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(dataSourceName, "name", updatedName),
				),
			},
		},
	})
}

func testDataAddOnConfig(name string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_add_on" "test" {
  name    = hcp_waypoint_add_on.test.name
}`, testAddOnConfig(name))
}
