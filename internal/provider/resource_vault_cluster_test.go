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
	hvn_id           = "test-hvn"
	cloud_provider   = "aws"
	region           = "us-west-2"
}

resource "hcp_vault_cluster" "test" {
	cluster_id       = "test-vault-cluster"
	hvn_id           = hcp_hvn.test.hvn_id
	tier             = "dev"
}

data "hcp_vault_cluster" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}

resource "hcp_vault_cluster_admin_token" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}
`

// This includes tests against both the resource, the corresponding datasource, and the dependent admin token resource
// to shorten testing time.
func TestAccVaultCluster(t *testing.T) {
	vaultClusterResourceName := "hcp_vault_cluster.test"
	vaultClusterDataSourceName := "data.hcp_vault_cluster.test"
	adminTokenResourceName := "hcp_vault_cluster_admin_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, false) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			// This step tests Vault cluster and admin token resource creation.
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(vaultClusterResourceName),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "tier", "DEV"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_version"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "project_id"),
					resource.TestCheckNoResourceAttr(vaultClusterResourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "created_at"),

					// Verifies admin token
					resource.TestCheckResourceAttr(adminTokenResourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttrSet(adminTokenResourceName, "token"),
					resource.TestCheckResourceAttrSet(adminTokenResourceName, "created_at"),
				),
			},
			// This step simulates an import of the resource.
			{
				ResourceName: vaultClusterResourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[vaultClusterResourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", vaultClusterResourceName)
					}

					return rs.Primary.Attributes["cluster_id"], nil
				},
				ImportStateVerify: true,
			},
			// This step is a subsequent terraform apply that verifies that no state is modified.
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(vaultClusterResourceName),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "cluster_id", "test-vault-cluster"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "hvn_id", "test-hvn"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "tier", "DEV"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "project_id"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_version"),
					resource.TestCheckNoResourceAttr(vaultClusterResourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "created_at"),
				),
			},
			// Tests datasource
			{
				Config: testConfig(testAccVaultClusterConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "cluster_id", vaultClusterDataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "hvn_id", vaultClusterDataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "public_endpoint", vaultClusterDataSourceName, "public_endpoint"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "min_vault_version", vaultClusterDataSourceName, "min_vault_version"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "tier", vaultClusterDataSourceName, "tier"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "organization_id", vaultClusterDataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "project_id", vaultClusterDataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "cloud_provider", vaultClusterDataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "region", vaultClusterDataSourceName, "region"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "namespace", vaultClusterDataSourceName, "namespace"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "vault_version", vaultClusterDataSourceName, "vault_version"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "vault_public_endpoint_url", vaultClusterDataSourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "vault_private_endpoint_url", vaultClusterDataSourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
					resource.TestCheckResourceAttrPair(vaultClusterResourceName, "created_at", vaultClusterDataSourceName, "created_at"),
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
