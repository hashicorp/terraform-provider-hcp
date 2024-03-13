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

func TestAccWaypointData_Application_Template_basic(t *testing.T) {
	// this is only used to verify the app template gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var appTemplateModel waypoint.ApplicationTemplateResourceModel
	resourceName := "hcp_waypoint_application_template.test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAppTemplateDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				// establish the base app template
				// note this reuses the config method from the app template
				// resource test
				Config: testAppTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAppTemplateExists(t, resourceName, &appTemplateModel),
				),
			},
			{
				// add a data source config to read the app template
				Config: testDataAppTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
				),
			},
			{
				// update the app template name, make sure it reflects in the data source
				Config: testDataAppTemplateConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(dataSourceName, "name", updatedName),
				),
			},
		},
	})
}

func testDataAppTemplateConfig(name string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_application_template" "test" {
  name    = hcp_waypoint_application_template.test.name
}`, testAppTemplateConfig(name))
}
