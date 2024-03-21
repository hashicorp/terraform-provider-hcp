// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccUserPrincipalDataSource(t *testing.T) {
	dataSourceAddress := "data.hcp_user_principal.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccUserPrincipalConfigEmailInput("cloud-experiences-tooling@hashicorp.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "user_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "email"),
				),
			},
			{
				Config: testAccUserPrincipalConfigUserIDInput("4e7b35b1-d4f5-419c-855b-a6034a33db54"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "user_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "email"),
				),
			},
			{
				Config: testAccUserPrincipalConfigBothInputs(
					"4e7b35b1-d4f5-419c-855b-a6034a33db54",
					"cloud-experiences-tooling@hashicorp.com",
				),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("Both email and user_id can not be set at the same time.")),
			},
			{
				Config:      testAccUserPrincipalConfigNoInputs(),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("Either user_id or email must be set in your input.")),
			},
		},
	})
}

func testAccUserPrincipalConfigEmailInput(email string) string {
	return fmt.Sprintf(`
	data "hcp_user_principal" "test" { 
		email = %q
	}
`, email)
}

func testAccUserPrincipalConfigUserIDInput(userID string) string {
	return fmt.Sprintf(`
	data "hcp_user_principal" "test" { 
		user_id = %q
	}
`, userID)
}

func testAccUserPrincipalConfigBothInputs(userID string, email string) string {
	return fmt.Sprintf(`
	data "hcp_user_principal" "test" { 
		user_id = %q
		email = %q
	}
`, userID, email)
}

func testAccUserPrincipalConfigNoInputs() string {
	return `
	data "hcp_user_principal" "test" { 
	}
	`
}
