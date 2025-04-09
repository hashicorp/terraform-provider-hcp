// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAcc_Waypoint_Data_Template_basic(t *testing.T) {
	t.Parallel()

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
					resource.TestCheckResourceAttr(dataSourceName, "terraform_no_code_module_id", "nocode-7ZQjQoaPXvzs6Hvp"),
					resource.TestCheckResourceAttr(dataSourceName, "terraform_execution_mode", "remote"),
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

func TestAcc_Waypoint_Data_template_with_variable_options(t *testing.T) {
	t.Parallel()

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
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.0.name", "faction"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.0.variable_type", "string"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.1.name", "vault_dweller_name"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.1.variable_type", "string"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.2.name", "vault_dweller_shelter"),
					resource.TestCheckResourceAttr(dataSourceName, "variable_options.2.variable_type", "string"),
				),
			},
		},
	})
}

func TestAcc_Waypoint_Data_template_with_actions(t *testing.T) {
	t.Parallel()

	var (
		appTemplateModel waypoint.TemplateResourceModel
		actionCfgModel   waypoint.ActionResourceModel
	)
	resourceName := "hcp_waypoint_template.actions_template_test"
	actionResourceName := "hcp_waypoint_action.test"
	dataSourceName := "data." + resourceName
	name := generateRandomName()
	actionName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			if err := testAccCheckWaypointTemplateDestroy(t, &appTemplateModel)(s); err != nil {
				return err
			}
			if err := testAccCheckWaypointActionDestroy(t, &actionCfgModel)(s); err != nil {
				return err
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				// establish the base template
				Config: testTemplateWithActionsConfig(name, actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTemplateExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointActionExists(t, actionResourceName, &actionCfgModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataAppTemplateWithAction(name, actionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "1"),
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

func testDataAppTemplateWithAction(templateName, actionName string) string {
	return fmt.Sprintf(`%s
data "hcp_waypoint_template" "actions_template_test" {
  name    = hcp_waypoint_template.actions_template_test.name
}`, testTemplateWithActionsConfig(templateName, actionName))
}
