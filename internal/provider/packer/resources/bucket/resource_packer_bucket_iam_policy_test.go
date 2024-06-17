// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bucket_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/testclient"
)

func TestAccPackerBucketIamPolicyResource(t *testing.T) {
	bucketName := "iam-policy-bucket"
	projectID := os.Getenv("HCP_PROJECT_ID")
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/contributor"
	roleName2 := "roles/admin"

	// A location is required to upsert the Packer Registry
	// Because of this we have to verify that the acceptance test value is set here, because the acceptance test check normally only occurs inside of resource.Test
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}
	loc := acctest.DefaultProjectLocation(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)
			testclient.UpsertRegistry(t, loc, nil)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccPackerBucketIamPolicy(projectName, roleName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_packer_bucket_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
			{
				Config: testAccPackerBucketIamPolicy(projectName, roleName2, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_packer_bucket_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccPackerBucketIamBindingResource(t *testing.T) {
	bucketName := "iam-binding-bucket"
	projectID := os.Getenv("HCP_PROJECT_ID")
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/contributor"

	// A location is required to upsert the Packer Registry
	// Because of this we have to verify that the acceptance test value is set here, because the acceptance test check normally only occurs inside of resource.Test
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}
	loc := acctest.DefaultProjectLocation(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)
			testclient.UpsertRegistry(t, loc, nil)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccPackerBucketIamBinding(projectName, roleName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_binding.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_packer_bucket_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccPackerBucketIamPolicy(projectName, roleName, bucketName string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = "test-sp"
	parent = %q
}

data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = %q
      principals = [
		hcp_service_principal.example.resource_id,
      ]
    },
  ]
}

resource "hcp_packer_bucket" "example" {
	name    = %q
}

resource "hcp_packer_bucket_iam_policy" "example" {
    resource_name = hcp_packer_bucket.example.resource_name
    policy_data = data.hcp_iam_policy.example.policy_data
}
`, projectName, roleName, bucketName)
}

func testAccPackerBucketIamBinding(projectName, roleName, bucketName string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = "hvs-sp"
	parent = %q
}

resource "hcp_packer_bucket" "example" {
	name    = %q
}

resource "hcp_packer_bucket_iam_binding" "example" {
	resource_name = hcp_packer_bucket.example.resource_name
	principal_id = hcp_service_principal.example.resource_id
	role         = %q
}
`, projectName, bucketName, roleName)
}
