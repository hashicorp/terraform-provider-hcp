// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestAccServicePrincipalResource_Project(t *testing.T) {
	spName := acctest.RandString(16)
	spNameUpdated := acctest.RandString(16)
	var sp models.HashicorpCloudIamServicePrincipal
	var sp2 models.HashicorpCloudIamServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalConfig(spName, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_service_principal.example", "name", spName),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "parent"),
					testAccServicePrincipalResourceExists(t, "hcp_service_principal.example", &sp),
					testAccCheckServicePrincipalValues(&sp, spName, true),
				),
			},
			{
				ResourceName:                         "hcp_service_principal.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccServicePrincipalImportID,
				ImportStateVerify:                    true,
			},
			{
				// Update the name
				Config: testAccServicePrincipalConfig(spNameUpdated, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_service_principal.example", "name", spNameUpdated),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_id"),
					testAccServicePrincipalResourceExists(t, "hcp_service_principal.example", &sp2),
					testAccCheckServicePrincipalValues(&sp2, spNameUpdated, true),
					func(_ *terraform.State) error {
						if sp.ID == sp2.ID {
							return fmt.Errorf("resource_ids match, indicating resource wasn't recreated")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccServicePrincipalResource_ExplicitProject(t *testing.T) {
	spName := acctest.RandString(16)
	var sp models.HashicorpCloudIamServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalConfig(spName, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_service_principal.example", "name", spName),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "parent"),
					testAccServicePrincipalResourceExists(t, "hcp_service_principal.example", &sp),
					testAccCheckServicePrincipalValues(&sp, spName, true),
				),
			},
		},
	})
}

func TestAccServicePrincipalResource_Organization(t *testing.T) {
	spName := acctest.RandString(16)
	var sp models.HashicorpCloudIamServicePrincipal

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalConfig(spName, false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hcp_service_principal.example", "name", spName),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_name"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "resource_id"),
					resource.TestCheckResourceAttrSet("hcp_service_principal.example", "parent"),
					testAccServicePrincipalResourceExists(t, "hcp_service_principal.example", &sp),
					testAccCheckServicePrincipalValues(&sp, spName, false),
				),
			},
			{
				ResourceName:                         "hcp_service_principal.example",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    testAccServicePrincipalImportID,
				ImportStateVerify:                    true,
			},
		},
	})
}

// testAccServicePrincipalImportID retrieves the resource_name so that it can be imported.
func testAccServicePrincipalImportID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["hcp_service_principal.example"]
	if !ok {
		return "", fmt.Errorf("resource not found")
	}

	id, ok := rs.Primary.Attributes["resource_name"]
	if !ok {
		return "", fmt.Errorf("resource_name not set")
	}

	return id, nil
}

func testAccServicePrincipalConfig(name string, explicitProject, explicitOrg bool) string {
	parentResource := ""
	parentParam := ""
	if explicitProject {
		parentResource = `data "hcp_project" "p" {}`
		parentParam = `parent = data.hcp_project.p.resource_name`
	} else if explicitOrg {
		parentResource = `data "hcp_organization" "o" {}`
		parentParam = `parent = data.hcp_organization.o.resource_name`
	}

	sp := fmt.Sprintf(`
resource "hcp_service_principal" "example" {
	name = %q
	%s
}`, name, parentParam)

	return fmt.Sprintf("%s\n%s", parentResource, sp)
}

// testAccCheckServicePrincipalResourceExists queries the API and retrieves the matching
// service principal.
func testAccServicePrincipalResourceExists(t *testing.T, resourceName string, sp *models.HashicorpCloudIamServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Get the SP resource name from state
		rname := rs.Primary.Attributes["resource_name"]

		// Fetch the SP
		client := acctest.HCPClients(t)
		getParams := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParams()
		getParams.ResourceName = rname
		res, err := client.ServicePrincipals.ServicePrincipalsServiceGetServicePrincipal(getParams, nil)
		if err != nil {
			return err
		}

		if res.GetPayload().ServicePrincipal == nil {
			return fmt.Errorf("ServicePrincipal(%s) not found", rname)
		}

		// assign the response project to the pointer
		*sp = *res.GetPayload().ServicePrincipal
		return nil
	}
}

func testAccCheckServicePrincipalValues(sp *models.HashicorpCloudIamServicePrincipal, name string, projectScoped bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sp.Name != name {
			return fmt.Errorf("bad name, expected \"%s\", got: %#v", name, sp.Name)
		}

		if projectScoped && sp.ProjectID == "" {
			return fmt.Errorf("expected project ID to be set")
		} else if !projectScoped && sp.ProjectID != "" {
			return fmt.Errorf("unexpected project ID: %s", sp.ProjectID)
		}

		return nil
	}
}
