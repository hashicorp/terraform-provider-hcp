// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccIAMPolicyDataSource(t *testing.T) {
	project := acctest.RandString(16)
	role1, role2 := "roles/viewer", "roles/admin"
	var p1, p2 models.HashicorpCloudResourcemanagerPolicy

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIAMPolicyConfig(project, role1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hcp_iam_policy.example", "policy_data"),
					testAccIAMPolicyData(t, "data.hcp_iam_policy.example", &p1),
					func(_ *terraform.State) error {

						// Check that the policy maps a single SP to the viewer
						// role
						if len(p1.Bindings) != 1 {
							return fmt.Errorf("unexpected number of bindings")
						}

						b := p1.Bindings[0]
						if b.RoleID != role1 {
							return fmt.Errorf("unexpected role")
						}

						return nil
					},
				),
			},
			{
				Config: testAccCheckIAMPolicyConfig(project, role2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hcp_iam_policy.example", "policy_data"),
					testAccIAMPolicyData(t, "data.hcp_iam_policy.example", &p2),
					func(_ *terraform.State) error {

						// Check that the policy maps a single SP to the viewer
						// role
						if len(p2.Bindings) != 1 {
							return fmt.Errorf("unexpected number of bindings")
						}

						b := p2.Bindings[0]
						if b.RoleID != role2 {
							return fmt.Errorf("unexpected role; got %q and want %q", b.RoleID, role2)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccIAMPolicyDataSource_Validation(t *testing.T) {

	numPrincipals := 2000
	principals := make([]string, numPrincipals)
	for i := 0; i < numPrincipals; i++ {
		principals[i] = fmt.Sprintf("%d", i)
	}

	resource.ParallelTest(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "hcp_iam_policy" "example" {
  bindings = [
	{
	  role = "roles/viewer"
	  principals = [
	    "1234"
	  ]
	},
	{
	  role = "roles/viewer"
	  principals = [
	    "5678"
	  ]
	},
  ]
}`,
				ExpectError: regexp.MustCompile(`(?m).*binding for role "roles/viewer" already defined.*`),
			},
			{
				Config: fmt.Sprintf(`
data "hcp_iam_policy" "example" {
  bindings = [
	{
	  role = "roles/viewer"
	  principals = [
	    %s
	  ]
	},
  ]
}`, strings.Join(principals, ", ")),
				ExpectError: regexp.MustCompile(`(?m).*A maximum of 1000 principals may be bound.*`),
			},
		},
	})
}

func testAccCheckIAMPolicyConfig(projectName, roleName string) string {
	return fmt.Sprintf(`
resource "hcp_project" "project" {
  name = %q
}

resource "hcp_service_principal" "sp" {
  name = "iam-acc-test"
  parent = hcp_project.project.resource_name
}

data "hcp_iam_policy" "example" {
  bindings = [
	{
	  role = %q
	  principals = [
	    hcp_service_principal.sp.resource_id
	  ]
	}
  ]
}
`, projectName, roleName)
}

func testAccIAMPolicyData(t *testing.T, resourceName string, policy *models.HashicorpCloudResourcemanagerPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Get the policy data
		pdata, ok := rs.Primary.Attributes["policy_data"]
		if !ok {
			return fmt.Errorf("Resource %q has no policy data set", resourceName)
		}

		var p models.HashicorpCloudResourcemanagerPolicy
		if err := p.UnmarshalBinary([]byte(pdata)); err != nil {
			return err
		}

		*policy = p
		return nil
	}
}
