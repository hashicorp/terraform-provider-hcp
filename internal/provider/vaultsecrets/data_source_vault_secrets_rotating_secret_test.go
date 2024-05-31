package vaultsecrets_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestDataSourceVaultSecretsRotatingSecret(t *testing.T) {
	t.Skip("skipping the test for now, we need to be able to create a rotating secret")
	testAppName := generateRandomSlug()
	//dataSourceAddress := "data.hcp_vault_secrets_rotating_secret.foo"

	testSecretName := "my_test_secret"
	testSecretValue := "insecurepassword"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, testAppName)

					createTestAppSecret(t, testAppName, testSecretName, "this shouldn't show up!")
					createTestAppSecret(t, testAppName, testSecretName, testSecretValue)
				},
			},
		},
	})
}
