package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVaultSecretsResourceSecret(t *testing.T) {
	testAppName := generateRandomSlug()
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					createTestApp(t, testAppName)
				},
				Config: fmt.Sprintf(`
				resource "hcp_vault_secrets_secret" "example" {
					app_name = %q
					secret_name = "test_secret"
					secret_value = "super secret"
				}`, testAppName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "app_name", testAppName),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_name", "test_secret"),
					resource.TestCheckResourceAttr("hcp_vault_secrets_secret.example", "secret_value", "super secret"),
				),
			},
		},
	})
}

const configResourceVaultSecretsSecret = `
resource "hcp_vault_secrets_secret" "example" {
  app_name = "%q"
  secret_name = "test_secret"
  secret_value = "super secret"
}
`
