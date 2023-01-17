package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	hvn1ResourceName           = "hcp_hvn.hvn1"
	hvn2ResourceName           = "hcp_hvn.hvn2"
	primaryVaultResourceName   = "hcp_vault_cluster.c1"
	secondaryVaultResourceName = "hcp_vault_cluster.c2"
)

func setTestAccPerformanceReplication_e2e(tfCode string) string {
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
`, tfCode)
}

func TestAccPerformanceReplication_Validations(t *testing.T) {
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
					tier         = lower(hcp_vault_cluster.c1.tier)
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
					tier         = lower(hcp_vault_cluster.c1.tier)
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
					tier         = lower(hcp_vault_cluster.c1.tier)
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
