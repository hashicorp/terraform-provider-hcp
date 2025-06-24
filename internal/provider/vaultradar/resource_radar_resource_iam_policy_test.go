// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

const (
	resourceContributor = "roles/vault-radar.resource-contributor"
	resourceViewer      = "roles/vault-radar.resource-viewer"
)

func TestRadarResourceIAMPolicy(t *testing.T) {
	// Requires Radar project already setup
	// Requires least one resource set up and registered with HCP.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	hcpResourceName := os.Getenv("RADAR_HCP_RESOURCE_NAME")

	if projectID == "" || hcpResourceName == "" {
		t.Skip("HCP_PROJECT_ID and RADAR_HCP_RESOURCE_NAME must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: createRadarResourceIAMPolicyConfig(projectID, resourceContributor, hcpResourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_policy.test", "resource_name", hcpResourceName),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_resource_iam_policy.test", "etag"),
				),
			},
			// UPDATE role
			{
				Config: createRadarResourceIAMPolicyConfig(projectID, resourceViewer, hcpResourceName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_resource_iam_policy.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_policy.test", "resource_name", hcpResourceName),
					resource.TestCheckResourceAttrSet("hcp_vault_radar_resource_iam_policy.test", "etag"),
				),
			},
			// DELETE happens automatically.
		},
	})

}

func createRadarResourceIAMPolicyConfig(projectID, role, hcpResourceName string) string {
	return fmt.Sprintf(`
		# Create a dev group.
		resource "hcp_group" "group" {
		  display_name = "tf-radar-dev-group-0"
		  description  = "group created from TF"
		}

		# Add a policy for the group to access Vault Radar with the developer role.
		resource "hcp_project_iam_binding" "binding" {
		  project_id   = %q
		  principal_id = hcp_group.group.resource_id
		  role         = "roles/vault-radar.developer"
		}

		# Create a Vault Radar Resource IAM policy for the group.
		data "hcp_iam_policy" "policy" {
		  bindings = [{
			role = %q
			principals = [hcp_group.group.resource_id]
		  }]
		}

		# Create a Vault Radar Resource IAM policy for the hcp resource name.
		resource "hcp_vault_radar_resource_iam_policy" "test" {
		  resource_name = %q
		  policy_data = data.hcp_iam_policy.policy.policy_data
		}
		`, projectID, role, hcpResourceName)
}

func TestRadarResourceIAMBinding(t *testing.T) {
	// Requires Radar project already setup
	// Requires least one resource set up and registered with HCP.
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	hcpResourceName := os.Getenv("RADAR_HCP_RESOURCE_NAME")

	if projectID == "" || hcpResourceName == "" {
		t.Skip("HCP_PROJECT_ID and RADAR_HCP_RESOURCE_NAME must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: createRadarResourceIAMBindingConfig(projectID, resourceContributor, hcpResourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_binding.test", "resource_name", hcpResourceName),
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_binding.test", "role", resourceContributor),
				),
			},
			// UPDATE role
			{
				Config: createRadarResourceIAMBindingConfig(projectID, resourceViewer, hcpResourceName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_resource_iam_binding.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_binding.test", "resource_name", hcpResourceName),
					resource.TestCheckResourceAttr("hcp_vault_radar_resource_iam_binding.test", "role", resourceViewer),
				),
			},
			// DELETE happens automatically.
		},
	})

}

func createRadarResourceIAMBindingConfig(projectID, role, hcpResourceName string) string {
	return fmt.Sprintf(`
		# Create a dev group.
		resource "hcp_group" "group" {
		  display_name = "tf-radar-dev-group-0"
		  description  = "group created from TF"
		}
		
		# Add a policy for the group to access Vault Radar with the developer role.
		resource "hcp_project_iam_binding" "binding" {
		  project_id   = %q
		  principal_id = hcp_group.group.resource_id
		  role         = "roles/vault-radar.developer"
		}
			
		# Create a Vault Radar Resource IAM binding for the hcp resource name.
		resource "hcp_vault_radar_resource_iam_binding" "test" {
		  resource_name = %q
		    principal_id = hcp_group.group.resource_id
			role         =  %q

		}
		`, projectID, hcpResourceName, role)
}
