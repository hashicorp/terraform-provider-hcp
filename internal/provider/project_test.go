package provider

import (
	"fmt"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/stretchr/testify/assert"
)

//have a provider instance declared that (eventually) contains a defined, existing project ID
//using providerFactory is possible

var preConfigProjectID string

var testAccNoProjectResourceConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}
`, hvnUniqueID)

func Test_GetProjectID(t *testing.T) {
	tests := []struct {
		name        string
		resProjID   string
		clientProj  string
		expectedID  string
		expectedErr string
	}{
		{"resource only project defined", "proj1", "", "proj1", ""},
		{"provider only project defined", "", "proj2", "proj2", ""},
		{"resource and provider project defined", "proj1", "proj2", "proj1", ""},
		{"project not defined", "", "", "", fmt.Sprintf("Project ID not defined. Verify that project ID is set either in the provider or in the resource config.")},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			projID, err := GetProjectID(testCase.resProjID, testCase.clientProj)
			assert.Equal(t, testCase.expectedID, projID)

			if testCase.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.expectedErr)
			}
		})
	}
}

//This includes tests against HVNS that contain a project ID defined in either the resource, the config, or both
func TestAccMultiProject(t *testing.T) {
	resourceName := "hcp_hvn.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		//function signature here requires a map of string to func, as defined above
		ProviderFactories: providerFactories,
		CheckDestroy: func(t *terraform.State) error {
			testAccCheckHvnDestroy(t)
			testAccMultiProjectDestroy(t)
			return nil
		},
		Steps: []resource.TestStep{

			{
				PlanOnly: true,
				//preconfig generates a project id
				PreConfig: func() {

					client := testAccProvider.Meta().(*clients.Client)
					loc := &sharedmodels.HashicorpCloudLocationLocation{
						ProjectID: client.Config.ProjectID}

					preConfigProjectID = loc.ProjectID

				},
				//assigns generated project id to the provider
				Config: fmt.Sprintf(`
				provider "hcp" {
					project_id = %s
				}
				%s
			`, preConfigProjectID, testAccNoProjectResourceConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", preConfigProjectID),
				),
			},
		},
	})

}

func testAccMultiProjectDestroy(s *terraform.State) error {

	//deletes projects associated with test provider and resource

	return nil
}
