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

func TestRadarSecretManagerVaultDedicatedResource(t *testing.T) {
	// Requires Project already setup with Radar.
	// Requires a Service Account with an Admin role on the Project.
	// Requires a Radar agent on K8 to be running and connected to Radar
	// Requires a Vault Dedicated instance with an auth policy enabling Kubernetes auth method.
	// See: https://github.com/hashicorp/vault-scanning-and-insights-cli/blob/main/docs/agent/vault-integration/kubernetes.md
	// Requires the following environment variables to be set:
	projectID := os.Getenv("HCP_PROJECT_ID")
	vaultURL := os.Getenv("RADAR_SM_VAULT_DEDICATED_VAULT_URL")
	mountPath := os.Getenv("RADAR_SM_VAULT_DEDICATED_MOUNT_PATH")
	roleName := os.Getenv("RADAR_SM_VAULT_DEDICATED_ROLE_NAME")

	if projectID == "" || vaultURL == "" || mountPath == "" || roleName == "" {
		t.Skip("HCP_PROJECT_ID, RADAR_SM_VAULT_DEDICATED_VAULT_URL, RADAR_SM_VAULT_DEDICATED_MOUNT_PATH and " +
			"RADAR_SM_VAULT_DEDICATED_ROLE_NAME must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_secret_manager_vault_dedicated" "example" {
						project_id = %q
						vault_url = %q
					    kubernetes = {
						  mount_path = %q
						  role_name  = %q
					    }
					}				
				`, projectID, vaultURL, mountPath, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "vault_url", vaultURL),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "access_read_write", "false"),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "kubernetes.mount_path", mountPath),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "kubernetes.role_name", roleName),
				),
			},
			// UPDATE access_read_write to true
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_radar_secret_manager_vault_dedicated" "example" {
						project_id = %q
						vault_url = %q
				        access_read_write = true
					    kubernetes = {
						  mount_path = %q
						  role_name  = %q
					    }
					}				
				`, projectID, vaultURL, mountPath, roleName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hcp_vault_radar_secret_manager_vault_dedicated.example", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "project_id", projectID),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "vault_url", vaultURL),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "access_read_write", "true"),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "kubernetes.mount_path", mountPath),
					resource.TestCheckResourceAttr("hcp_vault_radar_secret_manager_vault_dedicated.example", "kubernetes.role_name", roleName),
				),
			},
			// DELETE happens automatically.
		},
	})
}
