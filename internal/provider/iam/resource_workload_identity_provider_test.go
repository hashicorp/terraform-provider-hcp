// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

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

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityProviderConfig(spName, accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_iam_workload_identity_provider.example", "aws.account_id", accountID),
					resource.TestCheckResourceAttrSet("hcp_iam_workload_identity_provider.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_iam_workload_identity_provider.example", "resource_id"),
				),
			},
			{
				ResourceName:                         "hcp_iam_workload_identity_provider.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccWorkloadIdentityProviderImportID,
				ImportStateVerify:                    true,
			},
			{
				// Update the account id and expect an inplace update
				Config: testAccWorkloadIdentityProviderConfig(spName, accountID2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_iam_workload_identity_provider.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_iam_workload_identity_provider.example", "aws.account_id", accountID2),
					resource.TestCheckResourceAttrSet("hcp_iam_workload_identity_provider.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_iam_workload_identity_provider.example", "resource_id"),
				),
			},
		},
	})
}

// testAccWorkloadIdentityProviderImportID retrieves the resource_name so that it can be imported.
func testAccWorkloadIdentityProviderImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_iam_workload_identity_provider.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_name"]
	if !ok {
		return "", fmt.Errorf("resource_name not set")
	}

	return id, nil
}

func testAccWorkloadIdentityProviderConfig(spName, accountID string) string {
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
