package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

const vaultCluster = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "test-vault-cluster"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "dev"
}
`

// sets public_endpoint to true
const updatedVaultClusterPublic = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "test-vault-cluster"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "dev"
	public_endpoint    = true
}
`

// changes tier
const updatedVaultClusterTier = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "test-vault-cluster"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "standard_small"
}
`

// changes tier and sets public_endpoint to true
const updatedVaultClusterTierAndPublic = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "test-vault-cluster"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "standard_medium"
	public_endpoint    = true
}
`

func setTestAccVaultClusterConfig(vaultCluster string) string {
	return fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id            = "test-hvn"
	cloud_provider    = "aws"
	region            = "us-west-2"
}

%s

data "hcp_vault_cluster" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}

resource "hcp_vault_cluster_admin_token" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}
`, vaultCluster)
}

// This includes tests against both the resource, the corresponding datasource, and the dependent admin token resource
// to shorten testing time.
func TestAccVaultCluster(t *testing.T) {
	vaultClusterResourceName := "hcp_vault_cluster.test"
	vaultClusterDataSourceName := "data.hcp_vault_cluster.test"
	adminTokenResourceName := "hcp_vault_cluster_admin_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			// This step tests Vault cluster and admin token resource creation.
			{
				Config: testConfig(setTestAccVaultClusterConfig(vaultCluster)),
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
				Config: testConfig(setTestAccVaultClusterConfig(vaultCluster)),
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
				Config: testConfig(setTestAccVaultClusterConfig(vaultCluster)),
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
			// This step verifies the successful update of "public_endpoint".
			{
				Config: testConfig(setTestAccVaultClusterConfig(updatedVaultClusterPublic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(vaultClusterResourceName),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "public_endpoint", "true"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_public_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_public_endpoint_url", "8200"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
				),
			},
			// This step verifies the successful update of "tier".
			{
				Config: testConfig(setTestAccVaultClusterConfig(updatedVaultClusterTier)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(vaultClusterResourceName),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "tier", "STANDARD_SMALL"),
				),
			},
			// This step verifies the successful update of both "tier" and "public_endpoint".
			{
				Config: testConfig(setTestAccVaultClusterConfig(updatedVaultClusterTierAndPublic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(vaultClusterResourceName),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "tier", "STANDARD_MEDIUM"),
					resource.TestCheckResourceAttr(vaultClusterResourceName, "public_endpoint", "true"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_public_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_public_endpoint_url", "8200"),
					resource.TestCheckResourceAttrSet(vaultClusterResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(vaultClusterResourceName, "vault_private_endpoint_url", "8200"),
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

func setTestAccPerformanceReplication_e2e(vaultCluster string) string {
	return fmt.Sprintf(`
resource "hcp_hvn" "hvn1" {
	hvn_id            = "test-perf-hvn-1"
	cidr_block        = "172.25.16.0/20"
	cloud_provider    = "aws"
	region            = "us-west-2"
}

resource "hcp_hvn" "hvn2" {
	hvn_id            = "test-perf-hvn-2"
	cidr_block        = "172.24.16.0/20"
	cloud_provider    = "aws"
	region            = "us-west-2"
}

%s
`, vaultCluster)
}

func TestAccPerformanceReplication_Validations(t *testing.T) {
	hvn1ResourceName := "hcp_hvn.hvn1"
	hvn2ResourceName := "hcp_hvn.hvn2"
	primaryVaultResourceName := "hcp_vault_cluster.c1"
	secondaryVaultResourceName := "hcp_vault_cluster.c2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(setTestAccPerformanceReplication_e2e("")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(hvn1ResourceName, "hvn_id", "test-perf-hvn-1"),
					resource.TestCheckResourceAttr(hvn1ResourceName, "cidr_block", "172.25.16.0/20"),
					resource.TestCheckResourceAttr(hvn2ResourceName, "hvn_id", "test-perf-hvn-2"),
					resource.TestCheckResourceAttr(hvn2ResourceName, "cidr_block", "172.24.16.0/20"),
				),
			},
			{
				// invalid primary link supplied
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id   = "test-primary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = "plus_small"
					primary_link = "something"
					public_endpoint = true
				}
				`)),
				ExpectError: regexp.MustCompile(`invalid primary_link supplied*`),
			},
			{
				// incorrectly specify a paths_filter on a non-secondary
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id   = "test-primary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = "plus_small"
					paths_filter = ["path/a"]
				}
				`)),
				ExpectError: regexp.MustCompile(`only performance replication secondaries may specify a paths_filter`),
			},
			{
				// create a plus tier cluster successfully
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "cluster_id", "test-primary"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "hvn_id", "test-perf-hvn-1"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", "PLUS_SMALL"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "public_endpoint", "true"),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "vault_version"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "project_id"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "self_link"),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(primaryVaultResourceName, "vault_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(primaryVaultResourceName, "created_at"),
				),
			},
			{
				// secondary cluster creation failed as tier doesn't match the tier of primary
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = "plus_medium"
					primary_link = hcp_vault_cluster.c1.self_link
				}
				`)),
				ExpectError: regexp.MustCompile(`a secondary's tier must match that of its primary`),
			},
			{
				// secondary cluster creation failed as primary link is invalid
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = "not-present"
				}
				`)),
				ExpectError: regexp.MustCompile(`invalid primary_link supplied url`),
			},
			{
				// secondary cluster creation failed as min_vault_version is specified.
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id        = "test-secondary"
					hvn_id            = hcp_hvn.hvn1.hvn_id
					tier              = hcp_vault_cluster.c1.tier
					primary_link      = hcp_vault_cluster.c1.self_link
					min_vault_version = "v1.0.1"
				}
				`)),
				ExpectError: regexp.MustCompile(`min_vault_version should either be unset or match the primary cluster's`),
			},
			{
				// secondary cluster created successfully (same hvn)
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = hcp_vault_cluster.c1.self_link
					paths_filter = ["path/a", "path/b"]
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "cluster_id", "test-secondary"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "hvn_id", "test-perf-hvn-1"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", "PLUS_SMALL"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.0", "path/a"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.1", "path/b"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "vault_version"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "project_id"),
					resource.TestCheckNoResourceAttr(secondaryVaultResourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "self_link"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(secondaryVaultResourceName, "vault_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "created_at"),
				),
			},
			{
				// update paths filter
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = hcp_vault_cluster.c1.self_link
					paths_filter = ["path/a", "path/c"]
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.0", "path/a"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.1", "path/c"),
				),
			},
			{
				// delete paths filter
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn1.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = hcp_vault_cluster.c1.self_link
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(secondaryVaultResourceName, "paths_filter.0"),
				),
			},
			{
				// secondary cluster created successfully (different hvn)
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_small"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn2.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = hcp_vault_cluster.c1.self_link
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					testAccCheckVaultClusterExists(secondaryVaultResourceName),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "cluster_id", "test-secondary"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "hvn_id", "test-perf-hvn-2"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", "PLUS_SMALL"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "namespace", "admin"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "vault_version"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "project_id"),
					resource.TestCheckNoResourceAttr(secondaryVaultResourceName, "vault_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "self_link"),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "vault_private_endpoint_url"),
					testAccCheckFullURL(secondaryVaultResourceName, "vault_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(secondaryVaultResourceName, "created_at"),
				),
			},
			{
				// successfully scale replication group
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_medium"
					public_endpoint = true
				}
				resource "hcp_vault_cluster" "c2" {
					cluster_id   = "test-secondary"
					hvn_id       = hcp_hvn.hvn2.hvn_id
					tier         = hcp_vault_cluster.c1.tier
					primary_link = hcp_vault_cluster.c1.self_link
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", "PLUS_MEDIUM"),
					resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", "PLUS_MEDIUM"),
				),
			},
			{
				// successfully disable replication
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "plus_medium"
					public_endpoint = true
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
				),
			},
			{
				// successfully scale out of the Plus tier
				Config: testConfig(setTestAccPerformanceReplication_e2e(`
				resource "hcp_vault_cluster" "c1" {
					cluster_id      = "test-primary"
					hvn_id          = hcp_hvn.hvn1.hvn_id
					tier            = "starter_small"
					public_endpoint = true
				}
				`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", "STARTER_SMALL"),
				),
			},
		},
	})
}
