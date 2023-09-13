package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_dataSourceVaultSecretsSecret(t *testing.T) {
	testAppName := generateRandomSlug()
	dataSourceAddress := "data.hcp_vault_secrets_secret.foo"

	testSecretName := "secret_one"
	testSecretValue := "some value"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testPreCheck(t) },
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, testAppName)

					createTestAppSecret(t, testAppName, testSecretName, "this shouldn't show up!")
					createTestAppSecret(t, testAppName, testSecretName, testSecretValue)
				},
				Config: fmt.Sprintf(`
					data "hcp_vault_secrets_secret" "foo" {
						app_name    = %q
						secret_name = %q
					}`, testAppName, testSecretName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceAddress, "organization_id"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "project_id"),
					resource.TestCheckResourceAttr(dataSourceAddress, "secret_value", testSecretValue),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			deleteTestAppSecret(t, testAppName, testSecretName)
			deleteTestApp(t, testAppName)
			return nil
		},
	})
}
