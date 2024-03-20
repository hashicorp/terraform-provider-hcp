package waypoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

func TestAccWaypoint_Add_On_basic(t *testing.T) {
	var addOnModel waypoint.AddOnResourceModel
	resourceName := "hcp_waypoint_add_on.test"
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDestroy(t, &addOnModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
					testAccCheckWaypointAddOnName(t, &addOnModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				Config: testAddOnConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnExists(t, resourceName, &addOnModel),
					testAccCheckWaypointAddOnName(t, &addOnModel, updatedName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
				),
			},
		},
	})
}

// simple attribute check on the add-on definition received from the API
func testAccCheckWaypointAddOnName(t *testing.T, addOnModel *waypoint.AddOnResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if addOnModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected add-on name to be %s, but got %s", nameValue, addOnModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointAddOnExists(t *testing.T, resourceName string, addOnModel *waypoint.AddOnResourceModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		client := acctest.HCPClients(t)
		// Get the project ID and ID from state
		projectID := rs.Primary.Attributes["project_id"]
		appTempID := rs.Primary.Attributes["id"]
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		// Fetch the add-on
		addOn, err := clients.GetAddOnDefinitionByID(context.Background(), client, loc, appTempID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name and ID for now
		addOnModel.Name = types.StringValue(addOn.Name)
		addOnModel.ID = types.StringValue(addOn.ID)

		return nil
	}
}

func testAccCheckWaypointAddOnDestroy(t *testing.T, addOnModel *waypoint.AddOnResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := addOnModel.ID.ValueString()
		projectID := client.Config.ProjectID
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		definition, err := clients.GetAddOnByID(context.Background(), client, loc, id)
		if err != nil {
			// expected
			if clients.IsResponseCodeNotFound(err) {
				return nil
			}
			return err
		}

		// fall through, we expect a not found above but if we get this far then
		// the test should fail
		if definition != nil {
			return fmt.Errorf("expected add-on to be destroyed, but it still exists")
		}

		return fmt.Errorf("Both add-on and error were nil in destroy check, this should not happen")
	}
}

func testAddOnConfig(name string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_add_on" "test" {
  name    = %q
  application = {
    name  = "some name"
    id = "some id"
  }
  definition = {
	name = "some name"
	id = "some id"
  }
}`, name)
}
