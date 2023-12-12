// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serviceprincipal_test

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
	spTFName := "hcp_service_principal.example"
	var sp models.HashicorpCloudIamServicePrincipal
	var sp2 models.HashicorpCloudIamServicePrincipal

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: newResourceTestConfig(spName, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(spTFName, "name", spName),
					resource.TestCheckResourceAttrSet(spTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(spTFName, "resource_id"),
					resource.TestCheckResourceAttrSet(spTFName, "parent"),
					checkGetServicePrincipalResource(t, spTFName, &sp),
					checkResourceValues(&sp, spName, true),
				),
			},
			{
				ResourceName:                         spTFName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    servicePrincipalImportID(spTFName),
				ImportStateVerify:                    true,
			},
			{
				// Update the name
				Config: newResourceTestConfig(spNameUpdated, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(spTFName, "name", spNameUpdated),
					resource.TestCheckResourceAttrSet(spTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(spTFName, "resource_id"),
					checkGetServicePrincipalResource(t, spTFName, &sp2),
					checkResourceValues(&sp2, spNameUpdated, true),
					checkHasDifferentResourceID(&sp, &sp2),
				),
			},
		},
	})
}

func TestAccServicePrincipalResource_ExplicitProject(t *testing.T) {
	spName := acctest.RandString(16)
	spTFName := "hcp_service_principal.example"
	var sp models.HashicorpCloudIamServicePrincipal

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: newResourceTestConfig(spName, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(spTFName, "name", spName),
					resource.TestCheckResourceAttrSet(spTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(spTFName, "resource_id"),
					resource.TestCheckResourceAttrSet(spTFName, "parent"),
					checkGetServicePrincipalResource(t, spTFName, &sp),
					checkResourceValues(&sp, spName, true),
				),
			},
		},
	})
}

func TestAccServicePrincipalResource_Organization(t *testing.T) {
	spName := acctest.RandString(16)
	spTFName := "hcp_service_principal.example"
	var sp models.HashicorpCloudIamServicePrincipal

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: newResourceTestConfig(spName, false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(spTFName, "name", spName),
					resource.TestCheckResourceAttrSet(spTFName, "resource_name"),
					resource.TestCheckResourceAttrSet(spTFName, "resource_id"),
					resource.TestCheckResourceAttrSet(spTFName, "parent"),
					checkGetServicePrincipalResource(t, spTFName, &sp),
					checkResourceValues(&sp, spName, false),
				),
			},
			{
				ResourceName:                         spTFName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "resource_name",
				ImportStateIdFunc:                    servicePrincipalImportID(spTFName),
				ImportStateVerify:                    true,
			},
		},
	})
}

// servicePrincipalImportID retrieves the resource_name so that it can be imported.
func servicePrincipalImportID(tfResourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[tfResourceName]
		if !ok {
			return "", fmt.Errorf("resource not found")
		}

		id, ok := rs.Primary.Attributes["resource_name"]
		if !ok {
			return "", fmt.Errorf("resource_name not set")
		}

		return id, nil
	}
}

func newResourceTestConfig(name string, explicitProject, explicitOrg bool) string {
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

// checkGetServicePrincipalResource queries the API and retrieves the matching service principal.
func checkGetServicePrincipalResource(t *testing.T, resourceName string, sp *models.HashicorpCloudIamServicePrincipal) resource.TestCheckFunc {
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

func checkResourceValues(sp *models.HashicorpCloudIamServicePrincipal, name string, projectScoped bool) resource.TestCheckFunc {
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

func checkHasDifferentResourceID(sp1, sp2 *models.HashicorpCloudIamServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sp1.ID == sp2.ID {
			return fmt.Errorf("resource_ids match, indicating resource wasn't recreated")
		}
		return nil
	}
}
