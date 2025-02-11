// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccVaultSecretsResourceApp(t *testing.T) {
	var (
		integrationName1 = generateRandomSlug()
		appName1         = generateRandomSlug()
		appName2         = generateRandomSlug()
		description1     = "my description 1"
		description2     = "my description 2"
		projSyncName     = generateRandomSlug()
		groupSyncName    = generateRandomSlug()
		gitLabToken      = checkRequiredEnvVarOrFail(t, "GITLAB_ACCESS_TOKEN")
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial app
			{
				Config: appConfig(appName1, description1),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName1, description1, nil)...,
				),
			},
			// Changing an immutable field causes a recreation
			{
				Config: appConfig(appName2, description1),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName2, description1, nil)...,
				),
			},
			// Changing mutable fields causes an update
			{
				Config: appConfig(appName2, description2),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName2, description2, nil)...,
				),
			},
			// Changing the sync_names causes an update
			{
				Config: fmt.Sprintf(`
					resource "hcp_vault_secrets_integration" "acc_test" {
						name = %q
						capabilities = ["SYNC"]
						provider_type = "gitlab"
						gitlab_access = {
							token = %q
						}
					}
					resource "hcp_vault_secrets_sync" "gitlab_proj_sync" {
						name = %q
						integration_name = hcp_vault_secrets_integration.acc_test.name
						gitlab_config = {
						    scope = "PROJECT"
						    project_id = "123456789"
						}
					}
					resource "hcp_vault_secrets_sync" "gitlab_group_sync" {
						name = %q
						integration_name = hcp_vault_secrets_integration.acc_test.name
						gitlab_config = {
						    scope = "GROUP"
						    project_id = "987654321"
						}
					}
					resource "hcp_vault_secrets_app" "acc_test_app" {
						app_name    = %q
						description = %q
						sync_names = [hcp_vault_secrets_sync.gitlab_sync.name]
					}
				`, integrationName1, gitLabToken, projSyncName, groupSyncName, appName2, description2),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName2, description2, []string{projSyncName, groupSyncName})...,
				),
			},
			// Deleting the app out of band causes a recreation
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.DeleteApp(&secret_service.DeleteAppParams{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
						Name:           appName2,
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: appConfig(appName2, description2),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName2, description2, nil)...,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Pre-existing app can be imported
			{
				PreConfig: func() {
					t.Helper()
					client := acctest.HCPClients(t)
					_, err := client.VaultSecrets.CreateApp(&secret_service.CreateAppParams{
						OrganizationID: client.Config.OrganizationID,
						ProjectID:      client.Config.ProjectID,
						Body: &models.SecretServiceCreateAppBody{
							Name:        appName2,
							Description: description2,
						},
					}, nil)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: appConfig(appName2, description2),
				Check: resource.ComposeTestCheckFunc(
					appCheckFunc(appName2, description2, nil)...,
				),
				ResourceName:  "hcp_vault_secrets_app.acc_test_app",
				ImportStateId: appName2,
				ImportState:   true,
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			if appExists(t, appName1) {
				return fmt.Errorf("test app %s was not destroyed", appName1)
			}
			if appExists(t, appName2) {
				return fmt.Errorf("test app %s was not destroyed", appName2)
			}
			return nil
		},
	})
}

func appConfig(appName, description string) string {
	return fmt.Sprintf(`
	resource "hcp_vault_secrets_app" "acc_test_app" {
      app_name    = %q
      description = %q
	  sync_names = []
   }`, appName, description)
}

func appCheckFunc(appName, description string, syncNames []string) []resource.TestCheckFunc {
	syncsTestFns := make([]resource.TestCheckFunc, 0, len(syncNames))
	for _, syncName := range syncNames {
		syncsTestFns = append(syncsTestFns, resource.TestCheckTypeSetElemAttr("hcp_vault_secrets_app.acc_test_app", "sync_names.*", syncName))
	}

	testFns := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_app.acc_test_app", "organization_id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_app.acc_test_app", "id"),
		resource.TestCheckResourceAttrSet("hcp_vault_secrets_app.acc_test_app", "resource_name"),
		resource.TestCheckResourceAttr("hcp_vault_secrets_app.acc_test_app", "project_id", os.Getenv("HCP_PROJECT_ID")),
		resource.TestCheckResourceAttr("hcp_vault_secrets_app.acc_test_app", "app_name", appName),
		resource.TestCheckResourceAttr("hcp_vault_secrets_app.acc_test_app", "description", description)}

	testFns = append(testFns, syncsTestFns...)
	return testFns
}

func appExists(t *testing.T, name string) bool {
	t.Helper()

	client := acctest.HCPClients(t)

	response, err := client.VaultSecrets.GetApp(
		secret_service.NewGetAppParamsWithContext(ctx).
			WithOrganizationID(client.Config.OrganizationID).
			WithProjectID(client.Config.ProjectID).
			WithName(name), nil)
	if err != nil && !clients.IsResponseForbidden(err) {
		t.Fatal(err)
	}

	return !clients.IsResponseCodeNotFound(err) && response != nil && response.Payload != nil && response.Payload.App != nil
}
