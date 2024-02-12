package waypoint_test

import (
	"errors"
	"fmt"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"
)

func TestAccWaypointTfcConfig_basic(t *testing.T) {
	var tfcConfig waypoint.TfcConfigResourceModel
	resourceName := "hcp_waypoint_tfc_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointTfcConfigDestroy(t, &tfcConfig),
		Steps: []resource.TestStep{
			{
				Config: testConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointTfcConfigExists(t, resourceName, &tfcConfig),
					resource.TestCheckResourceAttr(resourceName, "tfc_org_name", "waypoint-tfc-testing"),
				),
			},
		},
	})
}

func testAccCheckWaypointTfcConfigExists(t *testing.T, resourceName string, tfcConfig *waypoint.TfcConfigResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID from state
		projectID := rs.Primary.Attributes["project_id"]
		orgID := client.Config.OrganizationID
		tfcConfig.ProjectID = types.StringValue(projectID)

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the project

		namespaceParams := &waypoint_service.WaypointServiceGetNamespaceParams{
			LocationOrganizationID: loc.OrganizationID,
			LocationProjectID:      loc.ProjectID,
		}
		// get namespace
		ns, err := client.Waypoint.WaypointServiceGetNamespace(namespaceParams, nil)
		if err != nil {
			return err
		}

		namespace := ns.GetPayload().Namespace
		params := &waypoint_service.WaypointServiceGetTFCConfigParams{
			NamespaceID: namespace.ID,
		}
		config, err := client.Waypoint.WaypointServiceGetTFCConfig(params, nil)
		if err != nil {
			return err
		}
		if config.Payload == nil || config.Payload.TfcConfig == nil {
			return errors.New("empty TFC Config payload")
		}
		return nil
	}
}

func testAccCheckWaypointTfcConfigDestroy(t *testing.T, tfcConfig *waypoint.TfcConfigResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		projectID := tfcConfig.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		namespaceParams := &waypoint_service.WaypointServiceGetNamespaceParams{
			LocationOrganizationID: orgID,
			LocationProjectID:      projectID,
		}
		// get namespace
		ns, err := client.Waypoint.WaypointServiceGetNamespace(namespaceParams, nil)
		if err != nil {
			return err
		}

		namespace := ns.GetPayload().Namespace
		params := &waypoint_service.WaypointServiceGetTFCConfigParams{
			NamespaceID: namespace.ID,
		}

		// Fetch the config
		_, err = client.Waypoint.WaypointServiceGetTFCConfig(params, nil)
		if clients.IsResponseCodeNotFound(err) {
			// we expect the config to be gone
			return nil
		}

		return fmt.Errorf("expected TFC Config to be destroyed, but no expected error returned: %v", err)
	}
}

func testConfig() string {
	return `
provider "hcp" {}
resource "hcp_waypoint_tfc_config" "test" {
  token        = "some fake token"
  tfc_org_name = "waypoint-tfc-testing"
}`
}
