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

func TestAcc_Waypoint_Data_Add_On_basic(t *testing.T) {
	t.Parallel()

	// this is only used to verify the add-on gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test"
	dataSourceName := "data." + resourceName
	addOnName := generateRandomName()
	templateName := generateRandomName()
	appName := generateRandomName()
	defName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				// establish the base add-on
				// note this reuses the config method from the add-on
				// resource test
				Config: testAddOnConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
				),
			},
			{
				// add a data source config to read the add-on
				Config: testDataAddOnConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", addOnName),
				),
			},
		},
	})
}

func TestAcc_Waypoint_AddOn_DataSource_WithInputVars(t *testing.T) {
	t.Parallel()

	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test_var_opts"
	dataSourceName := "data." + resourceName
	addOnName := generateRandomName()
	templateName := generateRandomName()
	appName := generateRandomName()
	defName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnWithInputVarsConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
				),
			},
			{
				Config: testDataAddOnWithInputVarsConfig(templateName, appName, defName, addOnName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", addOnName),
					resource.TestCheckResourceAttr(dataSourceName, "input_variables.#", "5"),
				),
			},
		},
	})
}

func testDataAddOnConfig(templateName string, appName string, defName string, addOnName string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_add_on" "test" {
  name    = hcp_waypoint_add_on.test.name
}`, testAddOnConfig(templateName, appName, defName, addOnName))
}

func testDataAddOnWithInputVarsConfig(templateName string, appName string, defName string, addOnName string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_add_on" "test_var_opts" {
  name    = hcp_waypoint_add_on.test_var_opts.name
}`, testAddOnWithInputVarsConfig(templateName, appName, defName, addOnName))
}
