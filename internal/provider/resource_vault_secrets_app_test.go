package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVaultSecretsResourceApp(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configResourceVaultSecretsApp,
				Check:  resource.TestCheckResourceAttr("hcp_vault_secrets_app.example", "app_name", "acctest-tf-app"),
			},
		},
	})
}

const configResourceVaultSecretsApp = `
resource "hcp_vault_secrets_app" "example" {
  app_name = "acctest-tf-app"
  description = "Acceptance test run"
}
`
