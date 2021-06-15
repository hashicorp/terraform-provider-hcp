package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var testAccVaultClusterConfig = `
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
	cluster_id            = "test-vault-cluster"
	hvn_id                = hcp_hvn.test.hvn_id
	tier									= "dev"
}

data "hcp_vault_cluster" "test" {
	cluster_id = hcp_vault_cluster.test.cluster_id
}
`

// This includes tests against both the resource and the corresponding datasource
// to shorten testing time.
func TestAccVaultCluster(t *testing.T) {
	resourceName := "hcp_vault_cluster.test"
	dataSourceName := "data.hcp_vault_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			// Tests create
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(resourceName, "tier", "DEV"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "vault_version"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckNoResourceAttr(resourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "vault_private_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			// This step simulates an import of the resource.
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					return rs.Primary.Attributes["cluster_id"], nil
				},
				ImportStateVerify: true,
			},
			// This step is a subsequent terraform apply that verifies that no state is modified.
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(resourceName, "tier", "DEV"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "vault_version"),
					resource.TestCheckNoResourceAttr(resourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "vault_private_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "public_endpoint", dataSourceName, "public_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "min_vault_version", dataSourceName, "min_vault_version"),
					resource.TestCheckResourceAttrPair(resourceName, "tier", dataSourceName, "tier"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace", dataSourceName, "namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_version", dataSourceName, "vault_version"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_public_endpoint_url", dataSourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "vault_private_endpoint_url", dataSourceName, "vault_private_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
				),
			},
		},
	})
}

func testAccCheckVaultClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*clients.Client)

		link, err := buildLinkFromURL(id, VaultClusterResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		clusterID := link.ID
		loc := link.Location

		if _, err := clients.GetVaultClusterByID(context.Background(), client, loc, clusterID); err != nil {
			return fmt.Errorf("unable to read Vault cluster %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckVaultClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_vault_cluster":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, VaultClusterResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			clusterID := link.ID
			loc := link.Location

			_, err = clients.GetVaultClusterByID(context.Background(), client, loc, clusterID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Vault cluster %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
