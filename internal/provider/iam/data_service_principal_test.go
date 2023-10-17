package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccServicePrincipalDataSource(t *testing.T) {
	name := acctest.RandString(16)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServicePrincipalConfig(name),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckDataSourceStateMatchesResourceStateWithIgnores(
						"data.hcp_service_principal.example",
						"hcp_service_principal.example",
						map[string]struct{}{
							"parent": {},
						},
					),
				),
			},
		},
	})
}

func testAccCheckServicePrincipalConfig(name string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
  name        = %q
}

data "hcp_service_principal" "example" {
  resource_name = hcp_service_principal.example.resource_name
}`, name)
}
