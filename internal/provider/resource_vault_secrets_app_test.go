// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVaultSecretsResourceApp(t *testing.T) {
	testAppName := generateRandomSlug()
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_app" "example" {
						app_name = %q
						description = "Acceptance test run"
				  }`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "app_name", testAppName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "description", "Acceptance test run"),
				),
			},
		},
	})
}
