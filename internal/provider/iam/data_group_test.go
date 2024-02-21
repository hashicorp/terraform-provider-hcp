package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

// TODO : update the below tests to use a created group resource via Terraform
func TestAccGroupDataSource(t *testing.T) {
	dataSourceAddress := "data.hcp_group.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig("int-tooling-e2e-test-group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "resource_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "description"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "display_name"),
				),
			},
		},
	})
}

func TestAccGroupDataSourceFullResourceName(t *testing.T) {
	dataSourceAddress := "data.hcp_group.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig("iam/organization/d11d7309-5072-44f9-aaea-c8f37c09a8b5/group/int-tooling-e2e-test-group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "resource_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "description"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "display_name"),
				),
			},
		},
	})
}

func testAccGroupConfig(resourceName string) string {
	return fmt.Sprintf(`
	data "hcp_group" "test" { 
		resource_name = %q
	}
`, resourceName)
}
