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

func TestAccWaypointData_Add_On_Definition_basic(t *testing.T) {
	// this is only used to verify the add-on definition gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var addOnDefinitionModel waypoint.AddOnDefinitionResourceModel
	resourceName := "hcp_waypoint_add_on_definition.test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDefinitionDestroy(t, &addOnDefinitionModel),
		Steps: []resource.TestStep{
			{
				// establish the base add-on definition
				// note this reuses the config method from the add-on definition
				// resource test
				Config: testAddOnDefinitionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnDefinitionExists(t, resourceName, &addOnDefinitionModel),
				),
			},
			{
				// add a data source config to read the add-on definition
				Config: testDataAddOnDefinitionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "terraform_execution_mode", "remote"),
				),
			},
			{
				// update the add-on definition name, make sure it reflects in the data source
				Config: testDataAddOnDefinitionConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(dataSourceName, "name", updatedName),
				),
			},
		},
	})
}

func testDataAddOnDefinitionConfig(name string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_add_on_definition" "test" {
  name    = hcp_waypoint_add_on_definition.test.name
}`, testAddOnDefinitionConfig(name))
}
