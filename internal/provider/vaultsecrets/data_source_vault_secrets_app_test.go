// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAcc_dataSourceVaultSecretsAppMigration(t *testing.T) {
	testAppName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_app.example"

	firstSecretName := "secret_one"
	secondSecretName := "secret_two"
	firstSecretValue := "hey, this is version 1!"
	secondSecretValue := "hey, this is version 2!"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create two secrets, one with an additional version and check the latest secrets from data source
			{
				PreConfig: func() {
					createTestApp(t, testAppName)

					createTestAppSecret(t, testAppName, firstSecretName, "this shouldn't show up!")
					createTestAppSecret(t, testAppName, firstSecretName, firstSecretValue)
					createTestAppSecret(t, testAppName, secondSecretName, secondSecretValue)
				},
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcp": {
						VersionConstraint: "~> 0.66.0",
						Source:            "hashicorp/hcp",
					},
				},
				Config: fmt.Sprintf(`
					data "hcp_vault_secrets_app" "example" {
						app_name    = %q
					}`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					resource.TestCheckResourceAttr(dataSourceAddress, "secrets.secret_one", firstSecretValue),
					resource.TestCheckResourceAttr(dataSourceAddress, "secrets.secret_two", secondSecretValue),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config: fmt.Sprintf(`
				data "hcp_vault_secrets_app" "example" {
					app_name    = %q
				}`, testAppName),
				PlanOnly: true,
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			deleteTestAppSecret(t, testAppName, firstSecretName)
			deleteTestAppSecret(t, testAppName, secondSecretName)
			deleteTestApp(t, testAppName)
			return nil
		},
	})
}

func TestAcc_dataSourceVaultSecretsApp(t *testing.T) {
	testAppName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_app.foo"

	firstSecretName := "secret_one"
	secondSecretName := "secret_two"
	firstSecretValue := "hey, this is version 1!"
	secondSecretValue := "hey, this is version 2!"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create two secrets, one with an additional version and check the latest secrets from data source
			{
				PreConfig: func() {
					createTestApp(t, testAppName)

					createTestAppSecret(t, testAppName, firstSecretName, "this shouldn't show up!")
					createTestAppSecret(t, testAppName, firstSecretName, firstSecretValue)
					createTestAppSecret(t, testAppName, secondSecretName, secondSecretValue)
				},
				Config: fmt.Sprintf(`
					data "hcp_vault_secrets_app" "foo" {
						app_name    = %q
					}`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					resource.TestCheckResourceAttr(dataSourceAddress, "secrets.secret_one", firstSecretValue),
					resource.TestCheckResourceAttr(dataSourceAddress, "secrets.secret_two", secondSecretValue),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			deleteTestAppSecret(t, testAppName, firstSecretName)
			deleteTestAppSecret(t, testAppName, secondSecretName)
			deleteTestApp(t, testAppName)
			return nil
		},
	})
}

func TestAcc_VaultSecretsOpenAppSecretsPagination(t *testing.T) {
	testAppName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_app.foo"
	secretCount := 12

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create 12 secrets and validate that all are returned and not just the default v2 api page size 10
			{
				PreConfig: func() {
					createTestApp(t, testAppName)

					for i := 1; i <= secretCount; i++ {
						createTestAppSecret(t, testAppName, "secret"+fmt.Sprint(i), "value"+fmt.Sprint(i))
					}
				},
				Config: fmt.Sprintf(`
					data "hcp_vault_secrets_app" "foo" {
						app_name    = %q
					}`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					//default page size for OpenAppSecrets v2 api is 10, validate that all secrets (pages) are retrieved and not just 1st page (10)
					resource.TestCheckResourceAttr(dataSourceAddress, "secrets.%", fmt.Sprintf("%v", secretCount)),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			for i := 1; i <= secretCount; i++ {
				deleteTestAppSecret(t, testAppName, "secret"+fmt.Sprint(i))
			}
			deleteTestApp(t, testAppName)
			return nil
		},
	})
}

func createTestApp(t *testing.T, appName string) {
	t.Helper()

	client := acctest.HCPClients(t)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.CreateVaultSecretsApp(context.Background(), client, loc, appName, "app description")
	if err != nil {
		t.Fatal(err)
	}
}

func createTestAppSecret(t *testing.T, appName, secretName, secretValue string) {
	t.Helper()

	client := acctest.HCPClients(t)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.CreateVaultSecretsAppSecret(context.Background(), client, loc, appName, secretName, secretValue)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTestAppSecret(t *testing.T, appName, secretName string) {
	t.Helper()

	client := acctest.HCPClients(t)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	err := clients.DeleteVaultSecretsAppSecret(context.Background(), client, loc, appName, secretName)
	if err != nil {
		t.Error(err)
	}
}

func deleteTestApp(t *testing.T, appName string) {
	t.Helper()

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	err := clients.DeleteVaultSecretsApp(context.Background(), client, loc, appName)
	if err != nil {
		t.Error(err)
	}
}

// generateRandomSlug will create a valid randomized slug with a prefix
func generateRandomSlug() string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return "hcp-provider-acctest-" + string(b)
}
