package provider

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	assertpkg "github.com/stretchr/testify/assert"
)

// testAccProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProvider *schema.Provider

// testAccProviderConfigure ensures testAccProvider is only configured once
//
// The testAccPreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// testAccProvider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"hcp": func() (*schema.Provider, error) {
		return New()(), nil
	},
}

func init() {
	testAccProvider = New()()
}

func TestProvider(t *testing.T) {
	if err := New()().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// testAccPreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It ensures
// testing functions that attempt to call HCP APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func testAccPreCheck(t *testing.T, requiredCreds map[string]bool) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		if os.Getenv("HCP_CLIENT_ID") == "" {
			t.Fatal("HCP_CLIENT_ID must be set for acceptance tests")
		}

		if os.Getenv("HCP_CLIENT_SECRET") == "" {
			t.Fatal("HCP_CLIENT_SECRET must be set for acceptance tests")
		}

		if requiredCreds["aws"] {
			if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
				t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
			}

			if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
				t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
			}

			if os.Getenv("AWS_SESSION_TOKEN") == "" {
				t.Fatal("AWS_SESSION_TOKEN must be set for acceptance tests")
			}
		}

		if requiredCreds["azure"] {
			if os.Getenv("AZURE_TENANT_ID") == "" {
				t.Fatal("AZURE_TENANT_ID must be set for acceptance tests")
			}

			if os.Getenv("AZURE_SUBSCRIPTION_ID") == "" {
				t.Fatal("AZURE_SUBSCRIPTION_ID must be set for acceptance tests")
			}
		}

		testDiag := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if testDiag.HasError() {
			t.Fatalf("unexpected error, exiting test: %#v", testDiag)
		}
	})
}

func testConfig(res ...string) string {
	provider := `provider "hcp" {}`

	c := []string{provider}
	c = append(c, res...)
	return strings.Join(c, "\n")
}

// If project ID is not defined on the provider config, the provider
//project ID becomes the organization's oldest existing project
func TestDetermineOldestProject(t *testing.T) {

	assert := assertpkg.New(t)

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
			assert.Equal(testCase.expectedProjID, oldestProject.ID)

		})

	}
}
