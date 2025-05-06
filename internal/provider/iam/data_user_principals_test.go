// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccUserPrincipalsDataSource(t *testing.T) {
	resourceName := "data.hcp_user_principals.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPrincipalsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestCheckResourceAttrSet(resourceName, "users.0.user_id"),
					resource.TestCheckResourceAttrSet(resourceName, "users.0.email"),
				),
			},
		},
	})
}

func testAccUserPrincipalsDataSourceConfig() string {
	return `
data "hcp_user_principals" "test" {
}
`
}
