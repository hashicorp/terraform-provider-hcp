// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccProjectsDataSource(t *testing.T) {

	resourceName := "data.hcp_projects.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `data "hcp_projects" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "projects.#"),
					resource.TestCheckResourceAttrSet(resourceName, "projects.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "projects.0.resource_name"),
					resource.TestCheckResourceAttrSet(resourceName, "projects.0.resource_id"),
				),
			},
		},
	})
}
