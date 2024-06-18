// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccGroupIamPolicyResource(t *testing.T) {
	// Test values for our integration tests in int.
	// resourceNameSuffix := "group_iam_policy_terraform_resource_test"
	resourceName := "iam/organization/d11d7309-5072-44f9-aaea-c8f37c09a8b5/group/group_iam_policy_terraform_resource_test"
	roleName := "roles/iam.group-manager"
	principalID := "4a836041-72f5-442d-a52f-af9e69f5a7f0"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)

			// Reset the Group's IAM policy before running the tests
			cleanupIAMPolicy(t, resourceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGroupIamPolicy(resourceName, principalID, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_group_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
			{
				ResourceName:                         "hcp_group_iam_policy.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        resourceName,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccGroupIamPolicy(resourceName, principalID, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_group_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccGroupIamBindingResource(t *testing.T) {
	// Test values for our integration tests in int.
	resourceNameSuffix := "group_iam_binding_terraform_resource_test"
	resourceName := "iam/organization/d11d7309-5072-44f9-aaea-c8f37c09a8b5/group/group_iam_binding_terraform_resource_test"
	roleName := "roles/iam.group-manager"
	principalID := "4a836041-72f5-442d-a52f-af9e69f5a7f0"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck: func() {
			acctest.PreCheck(t)

			// Reset the Group's IAM policy before running the tests
			cleanupIAMPolicy(t, resourceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGroupIamBinding(resourceNameSuffix, principalID, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "role"),
				),
			},
			{
				Config: testAccGroupIamBinding(resourceName, principalID, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "name"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_group_iam_binding.example", "role"),
				),
			},
		},
	})
}

func testAccGroupIamPolicy(resourceName, principalID, roleName string) string {
	return fmt.Sprintf(`
data "hcp_group" "example" {
	resource_name = %q
}

data "hcp_user_principal" "example" {
	user_id = %q
}

data "hcp_iam_policy" "example" {
  bindings = [
    {
      role = %q
      principals = [
				data.hcp_user_principal.example.user_id,
      ]
    },
  ]
}

resource "hcp_group_iam_policy" "example" {
	name = data.hcp_group.example.resource_name
	policy_data = data.hcp_iam_policy.example.policy_data
}
`, resourceName, principalID, roleName)
}

func testAccGroupIamBinding(resourceName, principalID, roleName string) string {
	return fmt.Sprintf(`
data "hcp_group" "example" {
	resource_name = %q
}

data "hcp_user_principal" "example" {
	user_id = %q
}

resource "hcp_group_iam_binding" "example" {
	name = data.hcp_group.example.resource_name
	principal_id = data.hcp_user_principal.example.user_id
	role = %q
}
`, resourceName, principalID, roleName)
}

// Set the IAM policy for a group to empty.
func cleanupIAMPolicy(t *testing.T, resourceName string) {
	client := acctest.HCPClients(t)

	getParams := resource_service.NewResourceServiceGetIamPolicyParams()
	getParams.ResourceName = &resourceName
	getRes, err := client.ResourceService.ResourceServiceGetIamPolicy(getParams, nil)
	if err != nil {
		t.Fatalf("Cleanup: Failed to get existing IAM policy: %v", err)
	}

	params := resource_service.NewResourceServiceSetIamPolicyParams()
	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		ResourceName: resourceName,
		Policy: &models.HashicorpCloudResourcemanagerPolicy{
			Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{},
			Etag:     getRes.Payload.Policy.Etag,
		},
	}
	_, err = client.ResourceService.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		t.Fatalf("Cleanup: Failed to reset IAM policy: %v", err)
	}
}
