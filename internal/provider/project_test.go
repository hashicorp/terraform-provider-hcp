package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

//have a provider instance declared that (eventually) contains a defined, existing project ID
//using providerFactory is possible

var (
	projUniqueID = fmt.Sprintf("hcp-multiproj-%s", time.Now().Format("200601021504"))
)

var testAccNoProjectResourceConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "%[1]s"
	cloud_provider = "aws"
	region         = "us-west-2"
}
`, projUniqueID)

var projectUUID = ""

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactoryWithProjectID = map[string]func() (*schema.Provider, error){
	"hcp": func() (*schema.Provider, error) {

		//how can i create the needed meta instance that normally would come from the provider
		//the meta instance would need to be defined to create a new client
		client := testAccProvider.Meta().(*clients.Client)
		loc := &sharedmodels.HashicorpCloudLocationLocation{
			ProjectID: client.Config.ProjectID,
		}
		fmt.Printf("THE LOCATION PROJect ID IS:")
		fmt.Printf(loc.ProjectID)
		//define interface to be passed into SetMeta call on provider
		projectIDinfo := map[string]string{"project_id": "projy-baby"}

		var provider = New()()
		provider.SetMeta(projectIDinfo)

		//Option1: cannot access the provider instance schema object directly since this requires a *terraform.ProviderSchemaRequest as a parameter
		//therefore cannot mutate in place project id directly with this approach
		//providerSchema, _ = provider.GetSchema()
		//providerSchema["project_id"] = "proj-123"
		//Option2: this configurecontextfunc is used when provider instance is first initialized, seems to be incorrect approach
		//provider.ConfigureContextFunc

		return provider, fmt.Errorf("an error")
	},
}

// This includes tests against HVNS that contain a project ID defined in either the resource, the config, or both
func TestAccMultiProject(t *testing.T) {
	resourceName := "hcp_hvn.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		//function signature here requires a map of string to func, as defined above
		ProviderFactories: providerFactoryWithProjectID,
		CheckDestroy:      testAccMultiProjectDestroy,
		Steps: []resource.TestStep{

			{
				Config: testConfig(testAccNoProjectResourceConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHvnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", "78910"),
				),
			},
		},
	})

}

func testAccMultiProjectDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_hvn":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, HvnResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			hvnID := link.ID
			loc := link.Location

			_, err = clients.GetHvnByID(context.Background(), client, loc, hvnID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed HVN %q: %v", id, err)
			}

		default:
			continue
		}
	}

	//delete projects associated with provider and resource

	return nil
}
