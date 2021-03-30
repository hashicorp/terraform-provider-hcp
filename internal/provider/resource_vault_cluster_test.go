package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var (
	testAccVaultClusterConfig = fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id         = "test-hvn"
	cloud_provider = "aws"
	region         = "us-west-2"
}
	
resource "hcp_vault_cluster" "test" {
	cluster_id            = "test-vault-cluster"
	hvn_id                = hcp_hvn.test.hvn_id
}
`)
)

func TestAccVaultCluster(t *testing.T) {
	resourceName := "hcp_vault_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
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
