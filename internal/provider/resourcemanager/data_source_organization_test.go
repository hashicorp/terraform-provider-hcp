// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccOrganizationDataSource(t *testing.T) {
	dataSourceAddress := "data.hcp_organization.org"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `data "hcp_organization" "org" { }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "resource_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "resource_name"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "name"),
				),
			},
		},
	})
}
