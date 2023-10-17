package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/acctest"
)

func TestAccServicePrincipalKeyResource(t *testing.T) {
	spName := acctest.RandString(16)
	var spk, spk2, spk3 models.HashicorpCloudIamServicePrincipalKey

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalKeyConfig(spName, "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hcp_service_principal_key.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_service_principal_key.example", "client_id"),
					resource.TestCheckResourceAttrSet("hcp_service_principal_key.example", "client_secret"),
					resource.TestCheckResourceAttrSet("hcp_service_principal_key.example", "service_principal"),
					testAccServicePrincipalKeyResourceExists(t, "hcp_service_principal_key.example", &spk),
				),
			},
			{
				// Update the trigger to force a new SPK
				Config: testAccServicePrincipalKeyConfig(spName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccServicePrincipalKeyResourceExists(t, "hcp_service_principal_key.example", &spk2),
					func(_ *terraform.State) error {
						if spk.ClientID == spk2.ClientID {
							return fmt.Errorf("client_ids match, indicating resource wasn't recreated")
						}
						return nil
					},
				),
			},
			{
				// Delete the underlying SPK
				Config: testAccServicePrincipalKeyConfig(spName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccDeleteServicePrincipalKey(t, &spk2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Check it was recreated
				Config: testAccServicePrincipalKeyConfig(spName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccServicePrincipalKeyResourceExists(t, "hcp_service_principal_key.example", &spk3),
					func(_ *terraform.State) error {
						if spk2.ClientID == spk3.ClientID {
							return fmt.Errorf("client_ids match, indicating resource wasn't recreated")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccServicePrincipalKeyConfig(spName, triggerVal string) string {
	return fmt.Sprintf(`
resource "hcp_service_principal" "sp" {
	name = %q
}

resource "hcp_service_principal_key" "example" {
	service_principal = hcp_service_principal.sp.resource_name
	rotate_triggers = {
		key = %q
	}
}
`, spName, triggerVal)
}

// testAccCheckServicePrincipalKeyResourceExists queries the API and retrieves the matching
// service principal key.
func testAccServicePrincipalKeyResourceExists(t *testing.T, resourceName string, spk *models.HashicorpCloudIamServicePrincipalKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Get the SP resource name from state
		spname := rs.Primary.Attributes["service_principal"]
		spkname := rs.Primary.Attributes["resource_name"]

		// Fetch the SP
		client := acctest.HCPClients(t)
		getParams := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParams()
		getParams.ResourceName = spname
		res, err := client.ServicePrincipals.ServicePrincipalsServiceGetServicePrincipal(getParams, nil)
		if err != nil {
			return err
		}

		if res.GetPayload().ServicePrincipal == nil {
			return fmt.Errorf("ServicePrincipal(%s) not found", spname)
		}

		// Scan the keys
		for _, k := range res.GetPayload().Keys {
			if k.ResourceName == spkname {
				*spk = *k
				return nil
			}
		}

		return fmt.Errorf("ServicePrincipalKey(%s) not found", spkname)
	}
}

// testAccDeleteServicePrincipalKey uses the API and deletes the
// service principal key.
func testAccDeleteServicePrincipalKey(t *testing.T, spk *models.HashicorpCloudIamServicePrincipalKey) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := acctest.HCPClients(t)
		deleteParams := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalKeyParams()
		deleteParams.ResourceName2 = spk.ResourceName
		_, err := client.ServicePrincipals.ServicePrincipalsServiceDeleteServicePrincipalKey(deleteParams, nil)
		if err != nil {
			return err
		}

		return nil
	}
}
