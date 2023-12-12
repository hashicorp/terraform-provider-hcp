// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityprovider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccWorkloadIdentityProviderResource(t *testing.T) {
	accountID, accountID2 := "123456789012", "123456789098"
	spName := acctest.RandString(16)
	wipTFName := "hcp_iam_workload_identity_provider.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: newResourceConfig(spName, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(wipTFName, "aws.account_id", accountID),
					resource.TestCheckResourceAttrSet(wipTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(wipTFName, "resource_id"),
				),
			},
			{
				ResourceName:                         wipTFName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    workloadIdentityProviderImportID(wipTFName),
				ImportStateVerify:                    true,
			},
			{
				// Update the account id and expect an inplace update
				Config: newResourceConfig(spName, accountID2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(wipTFName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(wipTFName, "aws.account_id", accountID2),
					resource.TestCheckResourceAttrSet(wipTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(wipTFName, "resource_id"),
				),
			},
		},
	})
}

// workloadIdentityProviderImportID retrieves the resource_name so that it can be imported.
func workloadIdentityProviderImportID(tfResourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[tfResourceName]
		if !ok {
			return "", fmt.Errorf("resource not found")
		}

		id, ok := rs.Primary.Attributes["resource_name"]
		if !ok {
			return "", fmt.Errorf("resource_name not set")
		}

		return id, nil
	}
}

func newResourceConfig(spName, accountID string) string {
	config := `
resource "hcp_service_principal" "example" {
	name = %q
}

resource "hcp_iam_workload_identity_provider" "example" {
	service_principal = hcp_service_principal.example.resource_name
	name = "aws"
	description = "wif"
	conditional_access = "aws.arn == \"arn:aws:sts::%s:assumed-role/bar\""

	aws = {
		account_id = %q
	}
} `

	return fmt.Sprintf(config, spName, accountID, accountID)
}
