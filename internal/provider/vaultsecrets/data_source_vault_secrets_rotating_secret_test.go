// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func checkRequiredEnvVarOrFail(t *testing.T, varName string) string {
	value, exists := os.LookupEnv(varName)
	if !exists {
		t.Skipf("%s must be set to execute this test", varName)
	}
	return value
}

func TestAcc_dataSourceVaultSecretsRotatingSecret(t *testing.T) {

	mongodbAtlasPublicKey := checkRequiredEnvVarOrFail(t, "MONGODB_ATLAS_API_PUBLIC_KEY")
	mongodbAtlasPrivateKey := checkRequiredEnvVarOrFail(t, "MONGODB_ATLAS_API_PRIVATE_KEY")
	mongodbAtlasGroupID := checkRequiredEnvVarOrFail(t, "MONGODB_ATLAS_GROUP_ID")
	mongodbAtlasDBName := checkRequiredEnvVarOrFail(t, "MONGODB_ATLAS_DB_NAME")

	testAppName := generateRandomSlug()
	testIntegrationName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_rotating_secret.foo"

	testSecretName := "secret_one"

	tfconfig := fmt.Sprintf(`data "hcp_vault_secrets_rotating_secret" "foo" {
		app_name = %q
		secret_name = %q
	}`, testAppName, testSecretName)

	client := acctest.HCPClients(t)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, testAppName)
					ctx := context.Background()

					_, err := clients.CreateMongoDBAtlasRotationIntegration(ctx, client, loc, testIntegrationName, mongodbAtlasPublicKey, mongodbAtlasPrivateKey)
					if err != nil {
						t.Fatalf("could not create mongodb rotation integration: %v", err)
					}

					reqBody := secretmodels.SecretServiceCreateMongoDBAtlasRotatingSecretBody{
						SecretName:         testSecretName,
						IntegrationName:    testIntegrationName,
						RotationPolicyName: "built-in:30-days-2-active",
						SecretDetails: &secretmodels.Secrets20231128MongoDBAtlasSecretDetails{
							MongodbGroupID: mongodbAtlasGroupID,
							MongodbRoles: []*secretmodels.Secrets20231128MongoDBRole{
								{
									DatabaseName:   mongodbAtlasDBName,
									RoleName:       "read",
									CollectionName: "",
								},
							},
							MongodbScopes: nil,
						},
					}
					_, err = clients.CreateMongoDBAtlasRotatingSecret(ctx, client, loc, testAppName, &reqBody)
					if err != nil {
						t.Fatalf("could not create rotating mongodb atlas secret: %v", err)
					}

					// block until the secret is done
					timeout := time.AfterFunc(10*time.Minute, func() {
						t.Fatalf("timed out waiting for mongodb rotating secret to be created")
					})

					waitForSecret := func() {
						for {
							state, err := clients.GetRotatingSecretState(ctx, client, loc, testAppName, testSecretName)
							if err != nil {
								t.Fatalf("could not get rotating secret state: %v", err)
							}
							switch *state.Status {
							case secretmodels.Secrets20231128RotatingSecretStatusERRORED:
								t.Fatalf("error rotating secret: %q", state.ErrorMessage)
							case secretmodels.Secrets20231128RotatingSecretStatusWAITINGFORNEXTROTATION:
								timeout.Stop()
								t.Log("secret successfully rotated")
								return
							default:
								t.Log("waiting to check rotating secret state")
								time.Sleep(10 * time.Second)
							}
						}
					}

					waitForSecret()

				},
				Config: tfconfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					resource.TestCheckResourceAttr(dataSourceAddress, "secret_values.%", "2"), // required: check the number of elements in the map
					resource.TestCheckResourceAttr(dataSourceAddress, "app_name", testAppName),
					resource.TestCheckResourceAttr(dataSourceAddress, "secret_provider", "mongodb-atlas"),
				),
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			ctx := context.Background()
			err := clients.DeleteVaultSecretsAppSecret(ctx, client, loc, testAppName, testSecretName)
			if err != nil {
				return fmt.Errorf("could not delete rotating secret: %v", err)
			}

			err = clients.DeleteMongoDBAtlasRotationIntegration(ctx, client, loc, testIntegrationName)
			if err != nil {
				return fmt.Errorf("could not delete rotation integration: %v", err)
			}

			deleteTestApp(t, testAppName)

			return nil
		},
	})
}
