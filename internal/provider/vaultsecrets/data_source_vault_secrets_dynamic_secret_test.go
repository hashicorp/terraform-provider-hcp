// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"context"
	"fmt"
	"testing"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

var ctx = context.Background()

func TestAcc_dataSourceVaultSecretsDynamicSecret(t *testing.T) {
	integrationRoleArn := checkRequiredEnvVarOrFail(t, "AWS_INTEGRATION_ROLE_ARN")
	secretRoleArn := checkRequiredEnvVarOrFail(t, "AWS_SECRET_ROLE_ARN")

	appName := generateRandomSlug()
	integrationName := generateRandomSlug()
	secretName := "secret_one"

	dataSourceAddress := "data.hcp_vault_secrets_dynamic_secret.foo"
	tfconfig := fmt.Sprintf(`data "hcp_vault_secrets_dynamic_secret" "foo" {
		app_name = %q
		secret_name = %q
	}`, appName, secretName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, appName)
					createTestAwsIntegration(t, integrationName, integrationRoleArn)
					createTestAwsDynamicSecret(t, appName, integrationName, secretName, secretRoleArn)
				},
				Config: tfconfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					resource.TestCheckResourceAttr(dataSourceAddress, "app_name", appName),
					resource.TestCheckResourceAttr(dataSourceAddress, "secret_name", secretName),
					resource.TestCheckResourceAttr(dataSourceAddress, "secret_provider", "aws"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "secret_values.access_key_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "secret_values.secret_access_key"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "secret_values.session_token"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "secret_values.assumed_role_user_arn"),
				),
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			deleteTestAwsDynamicSecret(t, appName, secretName)
			deleteTestAwsIntegration(t, integrationName)
			deleteTestApp(t, appName)

			return nil
		},
	})
}

func createTestAwsIntegration(t *testing.T, name, roleArn string) {
	t.Helper()

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.CreateAwsIntegration(ctx, client, loc, name, roleArn)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTestAwsIntegration(t *testing.T, name string) {
	t.Helper()

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	err := clients.DeleteAwsIntegration(ctx, client, loc, name)
	if err != nil {
		t.Error(err)
	}
}

func createTestAwsDynamicSecret(t *testing.T, appName, secretName, integrationName, roleArn string) {
	t.Helper()

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.CreateAwsDynamicSecret(ctx, client, loc, appName, secretName, integrationName, roleArn)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTestAwsDynamicSecret(t *testing.T, appName, secretName string) {
	t.Helper()

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	err := clients.DeleteAwsDynamicSecret(ctx, client, loc, appName, secretName)
	if err != nil {
		t.Error(err)
	}
}
