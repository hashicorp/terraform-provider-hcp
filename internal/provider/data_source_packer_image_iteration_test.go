package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	testAccPackerAlpineProductionImage = `
	data "hcp_packer_image_iteration" "alpine" {
		bucket  = "alpine"
		channel = "production"
	}`
)

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "data.hcp_packer_image_iteration.alpine"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.

			{
				PreConfig: func() {
					client := testAccProvider.Meta().(*clients.Client)
					_ = client
				},
				Config: testConfig(testAccPackerAlpineProductionImage),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
				),
			},
		},
	})
}
