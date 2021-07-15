package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	testAccPackerAlpineProductionImage = `
	data "hcp_packer_image" "alpine" {
		bucket  = "alpine"
		channel = "production"
	}`
)

func TestAcc_dataSourcePacker(t *testing.T) {
	resourceName := "hcp_packer_image.alpine"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,

		Steps: []resource.TestStep{
			// testing that getting the production channel of the alpine image
			// works.

			{
				Config: testConfig(testAccPackerAlpineProductionImage),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}
