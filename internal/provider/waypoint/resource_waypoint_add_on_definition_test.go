package waypoint_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/waypoint"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

func TestAccWaypoint_Add_On_Definition_basic(t *testing.T) {
	var appTemplateModel waypoint.AddOnDefinitionResourceModel
	resourceName := "hcp_waypoint_add_on_definition.test"
	name := generateRandomName()
	updatedName := generateRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWaypointAddOnDefinitionDestroy(t, &appTemplateModel),
		Steps: []resource.TestStep{
			{
				Config: testAddOnDefinitionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnDefinitionExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointAddOnDefinitionName(t, &appTemplateModel, name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				Config: testAddOnDefinitionConfig(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWaypointAddOnDefinitionExists(t, resourceName, &appTemplateModel),
					testAccCheckWaypointAddOnDefinitionName(t, &appTemplateModel, updatedName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
				),
			},
		},
	})
}

// simple attribute check on the add-on definition received from the API
func testAccCheckWaypointAddOnDefinitionName(t *testing.T, addOnDefinitionModel *waypoint.AddOnDefinitionResourceModel, nameValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if addOnDefinitionModel.Name.ValueString() != nameValue {
			return fmt.Errorf("expected add-on definition name to be %s, but got %s", nameValue, addOnDefinitionModel.Name.ValueString())
		}
		return nil
	}
}

func testAccCheckWaypointAddOnDefinitionExists(t *testing.T, resourceName string, definitionModel *waypoint.AddOnDefinitionResourceModel) resource.TestCheckFunc {
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

		// Fetch the add-on definition
		definition, err := clients.GetAddOnDefinitionByID(context.Background(), client, loc, appTempID)
		if err != nil {
			return err
		}

		// at this time we're only verifing existence and not checking all the
		// values, so only set name,id, and project id for now
		definitionModel.Name = types.StringValue(definition.Name)
		definitionModel.ID = types.StringValue(definition.ID)
		definitionModel.ProjectID = types.StringValue(projectID)

		return nil
	}
}

func testAccCheckWaypointAddOnDefinitionDestroy(t *testing.T, definitionModel *waypoint.AddOnDefinitionResourceModel) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		id := definitionModel.ID.ValueString()
		projectID := definitionModel.ProjectID.ValueString()
		orgID := client.Config.OrganizationID

		loc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: orgID,
			ProjectID:      projectID,
		}

		definition, err := clients.GetAddOnDefinitionByID(context.Background(), client, loc, id)
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
			return fmt.Errorf("expected add-on definition to be destroyed, but it still exists")
		}

		return fmt.Errorf("both definition and error were nil in destroy check, this should not happen")
	}
}

// TODO: (Henry) Add remaining add-on definition fields to test (tags, labels, readmemarkdown, definition.. etc)
func testAddOnDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "hcp_waypoint_add_on_definition" "test" {
  name    = %q
  summary = "some summary for fun"
  description = "some description for fun"
  terraform_no_code_module = {
    source  = "some source"
    version = "some version"
  }
  terraform_cloud_workspace_details = {
    name                 = "some name"
    terraform_project_id = "some id"
  }
}`, name)
}

// generateRandomName will create a valid randomized name
// TODO: (Henry) This function is duplicated in multiple tests. It should be moved to a common location.
func generateRandomName() string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return "hcp-provider-acctest-" + string(b)
}
