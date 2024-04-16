// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	rmmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsAppIamPolicyResource(t *testing.T) {
	appName := generateRandomSlug()
	projectID := os.Getenv("HCP_PROJECT_ID")
	resourceName := fmt.Sprintf("secrets/project/%s/app/%s", projectID, appName)
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/contributor"
	roleName2 := "roles/viewer"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVaultSecretsAppIamPolicy(projectName, roleName, appName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_vault_secrets_app_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
			{
				Config: testAccVaultSecretsAppIamPolicy(projectName, roleName2, appName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "etag"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_policy.example", "policy_data"),
					resource.TestCheckResourceAttrPair("hcp_vault_secrets_app_iam_policy.example", "policy_data", "data.hcp_iam_policy.example", "policy_data"),
				),
			},
		},
	})
}

func TestAccVaultSecretsAppIamBindingResource(t *testing.T) {
	appName := generateRandomSlug()
	projectID := os.Getenv("HCP_PROJECT_ID")
	projectName := fmt.Sprintf("project/%s", projectID)
	roleName := "roles/contributor"
	resourceName := fmt.Sprintf("secrets/project/%s/app/%s", projectID, appName)
	spName := "test-sp"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() { setHVSIAMPolicy(t, appName, projectName, spName) },
				Config:    testAccVaultSecretsAppIamBinding(projectName, appName, resourceName, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "principal_id"),
					resource.TestCheckResourceAttrSet("hcp_vault_secrets_app_iam_binding.example", "role"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			deleteTestApp(t, appName)
			deleteServicePrincipal(t, projectName, spName)
			return nil
		},
	})
}

func testAccVaultSecretsAppIamPolicy(projectName, roleName, appName, resourceName string) string {
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

resource "hcp_vault_secrets_app" "example" {
	app_name    = %q
	description = "My new app!"
}

resource "hcp_vault_secrets_app_iam_policy" "example" {
    resource_name = %q
    policy_data = data.hcp_iam_policy.example.policy_data
}
`, projectName, roleName, appName, resourceName)
}

func testAccVaultSecretsAppIamBinding(projectName, appName, resourceName, roleName string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = "hvs-sp"
	parent = %q
}

resource "hcp_vault_secrets_app_iam_binding" "example" {
	resource_name = %q
	principal_id = hcp_service_principal.example.resource_id
	role         = %q
}
`, projectName, resourceName, roleName)
}

func setHVSIAMPolicy(t *testing.T, appName, projectResourceName, spName string) {
	t.Helper()

	createTestApp(t, appName)
	principalID := createServicePrincipal(t, projectResourceName, spName)

	client := acctest.HCPClients(t)

	params := resource_service.NewResourceServiceSetIamPolicyParams()
	spMemberType := rmmodels.HashicorpCloudResourcemanagerPolicyBindingMemberTypeSERVICEPRINCIPAL

	policy := &rmmodels.HashicorpCloudResourcemanagerPolicy{
		Bindings: []*rmmodels.HashicorpCloudResourcemanagerPolicyBinding{
			{
				RoleID: "roles/viewer",
				Members: []*rmmodels.HashicorpCloudResourcemanagerPolicyBindingMember{
					{
						MemberID:   principalID,
						MemberType: &spMemberType,
					},
				},
			},
		},
	}
	resourceName := fmt.Sprintf("secrets/%s/app/%s", projectResourceName, appName)
	params.Body = &rmmodels.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		Policy:       policy,
		ResourceName: resourceName,
	}

	_, err := client.ResourceService.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		if err != nil {
			t.Fatalf("unexpected error setting IAM policy: %v", err)
			return
		}
	}
}

func createServicePrincipal(t *testing.T, parentName, name string) string {
	client := acctest.HCPClients(t)

	params := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalParams()
	params.ParentResourceName = parentName
	params.Body = service_principals_service.ServicePrincipalsServiceCreateServicePrincipalBody{
		Name: name,
	}
	resp, err := client.ServicePrincipals.ServicePrincipalsServiceCreateServicePrincipal(params, nil)
	if err != nil {
		if err != nil {
			t.Fatalf("unexpected error creating service principal: %v", err)
			return ""
		}
	}
	return resp.GetPayload().ServicePrincipal.ID
}

func deleteServicePrincipal(t *testing.T, parentName, name string) {
	t.Helper()
	client := acctest.HCPClients(t)

	params := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalParams()
	params.ResourceName = fmt.Sprintf("iam/%s/service-principal/%s", parentName, name)

	_, err := client.ServicePrincipals.ServicePrincipalsServiceDeleteServicePrincipal(params, nil)
	if err != nil {
		if err != nil {
			t.Fatalf("unexpected error deleting service principal: %v", err)
			return
		}
	}
}
