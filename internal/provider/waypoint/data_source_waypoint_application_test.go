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

func TestAccWaypoint_Application_DataSource_basic(t *testing.T) {
	// this is only used to verify the app template gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var applicationModel waypoint.ApplicationResourceModel
	resourceName := "hcp_waypoint_application.test"
	dataSourceName := "data." + resourceName
	templateName := generateRandomName()
	applicationName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				// establish the base app template and application
				Config: testApplicationConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
				),
			},
			{
				// add a data source config to read the app template
				Config: testDataApplicationConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", applicationName),
				),
			},
		},
	})
}

func testDataApplicationConfig(templateName, applicationName string) string {
	return fmt.Sprintf(`%s

data "hcp_waypoint_application" "test" {
  name    = hcp_waypoint_application.test.name
}`, testApplicationConfig(templateName, applicationName))
}
