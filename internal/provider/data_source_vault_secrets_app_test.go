// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func TestAcc_dataSourceVaultSecretsApp(t *testing.T) {
	testAppName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_app.foo"

	firstSecretName := "secret_one"
	secondSecretName := "secret_two"
	firstSecretValue := "hey, this is version 1!"
	secondSecretValue := "hey, this is version 2!"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{}) },
		ProviderFactories: providerFactories,
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

func createTestApp(t *testing.T, appName string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.CreateVaultSecretsApp(context.Background(), client, loc, appName)
	if err != nil {
		t.Fatal(err)
	}
}

func createTestAppSecret(t *testing.T, appName, secretName, secretValue string) {
	t.Helper()

	client := testAccProvider.Meta().(*clients.Client)
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

	client := testAccProvider.Meta().(*clients.Client)
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

	client := testAccProvider.Meta().(*clients.Client)
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
