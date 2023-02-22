package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// Project ID is read from the resource config or provider
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
		{"project not defined", "", "", "", "project ID not defined. Verify that project ID is set either in the provider or in the resource config"},
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

// If project ID is not defined on the provider or resource config, the provider
// project ID becomes the organization's oldest existing project
func TestDetermineOldestProject(t *testing.T) {

	testCases := []struct {
		name           string
		projArray      []*models.HashicorpCloudResourcemanagerProject
		expectedProjID string
	}{
		{
			name: "One Project",
			projArray: []*models.HashicorpCloudResourcemanagerProject{
				{
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					ID:        "proj1",
				},
			},
			expectedProjID: "proj1",
		},
		{
			name: "Two Projects",
			projArray: []*models.HashicorpCloudResourcemanagerProject{
				{
					ID:        "proj1",
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj2",
					CreatedAt: strfmt.DateTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			expectedProjID: "proj2",
		},
		{
			name: "Three Projects",
			projArray: []*models.HashicorpCloudResourcemanagerProject{
				{
					ID:        "proj1",
					CreatedAt: strfmt.DateTime(time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj2",
					CreatedAt: strfmt.DateTime(time.Date(2007, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "proj3",
					CreatedAt: strfmt.DateTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			expectedProjID: "proj2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			oldestProject := getOldestProject(testCase.projArray)
			assert.Equal(t, testCase.expectedProjID, oldestProject.ID)

		})

	}
}

var (
	hvnIDUnique = fmt.Sprintf("hcp-provider-test-%s", time.Now().Format("200601021504"))
)

var providerProjectID = "prov-project-id-invalid"

func TestAccMultiProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(t *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{

			{
				PlanOnly: true,
				Config: fmt.Sprintf(`
				provider "hcp" {
					project_id = "%[1]s"
				}
				resource "hcp_hvn" "test" {
					hvn_id         = "%[2]s"
					cloud_provider = "aws"
					region         = "us-west-2"
				}
				`, providerProjectID, hvnIDUnique),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`unable to fetch project "%s": could not complete request: please ensure your HCP_API_HOST, HCP_CLIENT_ID, and HCP_CLIENT_SECRET are correct`, providerProjectID)),
			},
		},
	})

}

func TestAccMultiProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(t *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{

			{
				PlanOnly: true,
				Config: fmt.Sprintf(`
				provider "hcp" {}
				resource "hcp_hvn" "test" {
					hvn_id         = "%[1]s"
					project_id = "resource-project-id-invalid"
					cloud_provider = "aws"
					region         = "us-west-2"
				}
				`, hvnIDUnique),
				ExpectError: regexp.MustCompile("Invalid project ID provided for resource"),
			},
		},
	})

}
