package vaultsecrets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAcc_ephemeralVaultSecretsSecret(t *testing.T) {
	t.Parallel()

	testAppName := generateRandomSlug()
	testSecretName := "secret_one"
	testSecretValue := "some value"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			PreConfig: func() {
				createTestApp(t, testAppName)

				createTestAppSecret(t, testAppName, testSecretName, "this shouldn't show up!")
				createTestAppSecret(t, testAppName, testSecretName, testSecretValue)
			},
			Config: fmt.Sprintf(`
ephemeral "hcp_vault_secrets_secret" "bar" {
   app_name    = %q
   secret_name = %q
}`, testSecretName, testSecretValue),
		}},
	})
}
