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

func TestAccWaypointData_Template_basic(t *testing.T) {
	// this is only used to verify the template gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var appTemplateModel waypoint.TemplateResourceModel
	resourceName := "hcp_waypoint_template.test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointTemplateDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				// establish the base template
				// note this reuses the config method from the template
				// resource test
				Config: testTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataAppTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
				),
			},
			{
				// update the template name, make sure it reflects in the data source
				Config: testDataAppTemplateConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(dataSourceName, "name", updatedName),
				),
			},
		},
	})
}

func TestAccWaypointData_template_with_variable_options(t *testing.T) {
	// this is only used to verify the template gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var appTemplateModel waypoint.TemplateResourceModel
	resourceName := "hcp_waypoint_template.var_opts_test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointTemplateDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				// establish the base template
				// note this reuses the config method from the template
				// resource test
				Config: testTemplateConfigWithVarOpts(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataAppTemplateWithVariablesWithOptionsConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.0.name", "faction"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.0.variable_type", "string"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.1.name", "vault_dweller_name"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.1.variable_type", "string"),
				),
			},
		},
	})

}

func testDataAppTemplateConfig(name string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_template" "test" {
  name    = hcp_waypoint_template.test.name
}`, testTemplateConfig(name))
}

func testDataAppTemplateWithVariablesWithOptionsConfig(name string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_template" "var_opts_test" {
  name    = hcp_waypoint_template.var_opts_test.name
}`, testTemplateConfigWithVarOpts(name))
}
