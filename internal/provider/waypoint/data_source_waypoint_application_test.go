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

func TestAcc_Waypoint_Application_DataSource_basic(t *testing.T) {
	t.Parallel()

	// this is only used to verify the template gets cleaned up in the end
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
				// establish the base template and application
				Config: testApplicationConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataApplicationConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", applicationName),
				),
			},
		},
	})
}

func TestAcc_Waypoint_Application_DataSource_WithInputVars(t *testing.T) {
	t.Parallel()

	// this is only used to verify the template gets cleaned up in the end
	// of the test, and not used for any other purpose at this time
	var applicationModel waypoint.ApplicationResourceModel
	resourceName := "hcp_waypoint_application.test_var_opts"
	dataSourceName := "data." + resourceName
	templateName := generateRandomName()
	applicationName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointApplicationDestroy(t, &applicationModel),
		Steps: []resource.TestStep{
			{
				// establish the base template and application
				Config: testApplicationWithInputVarsConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataApplicationWithInputVarsConfig(templateName, applicationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", applicationName),
					resource.TestCheckResourceAttr(dataSourceName, "input_variables.#", "4"),
				),
			},
		},
	})
}

func TestAcc_Waypoint_Application_DataSource_WithActions(t *testing.T) {
	t.Parallel()

	var (
		appTemplateModel waypoint.TemplateResourceModel
		applicationModel waypoint.ApplicationResourceModel
		actionCfgModel   waypoint.ActionResourceModel
	)
	templateResourceName := "hcp_waypoint_template.actions_template_test"
	resourceName := "hcp_waypoint_application.actions_application_test"
	actionResourceName := "hcp_waypoint_action.test"
	dataSourceName := "data." + resourceName
	templateName := generateRandomName()
	applicationName := generateRandomName()
	actionName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			if err := testAccCheckWaypointApplicationDestroy(t, &applicationModel)(s); err != nil {
				return err
			}
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
				// establish the base template and app
				Config: testTemplateWithAppAndActionsConfig(templateName, applicationName, actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointApplicationExists(t, resourceName, &applicationModel),
					testAccCheckWaypointTemplateExists(t, templateResourceName, &appTemplateModel),
					testAccCheckWaypointActionExists(t, actionResourceName, &actionCfgModel),
				),
			},
			{
				// add a data source config to read the template
				Config: testDataApplicationWithAction(templateName, applicationName, actionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", applicationName),
					resource.TestCheckResourceAttr(dataSourceName, "actions.#", "1"),
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

func testDataApplicationWithInputVarsConfig(templateName, applicationName string) string {
	return fmt.Sprintf(`%s

data "hcp_waypoint_application" "test_var_opts" {
  name    = hcp_waypoint_application.test_var_opts.name
}`, testApplicationWithInputVarsConfig(templateName, applicationName))
}

func testDataApplicationWithAction(templateName, applicationName, actionName string) string {
	return fmt.Sprintf(`%s

data "hcp_waypoint_application" "actions_application_test" {
  name		= hcp_waypoint_application.actions_application_test.name
}`, testTemplateWithAppAndActionsConfig(templateName, applicationName, actionName))
}
