// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var consulClusterUniqueID = fmt.Sprintf("test-%s", time.Now().Format("200601021504"))
var consulClusterHVNUniqueID = fmt.Sprintf("test-hvn-%s", time.Now().Format("200601021504"))

var consulCluster = consulClusterConfig(consulClusterUniqueID, `
	ip_allowlist {
    	address = "172.25.16.0/24"
    	description = "this is an IPV4 address"
  	}
`)

var updatedConsulCluster = consulClusterConfig(consulClusterUniqueID, `
	size	   = "medium"

	ip_allowlist {
    	address = "172.25.14.0/24"
 		description = "this is an updated IPV4 address"
    }
`)

var createConsulClusterCIDRExceeded = consulClusterConfig("test-failure", `
	ip_allowlist {
		address = "172.25.16.0/24"
		description = "this is a first IPV4 address"
	}
	
	ip_allowlist {
		address = "172.25.10.0/24"
		description = "this is a secondIPV4 address"
	}

	ip_allowlist {
		address = "172.25.13.0/24"
		description = "this is a third IPV4 address"
	}

	ip_allowlist {
		address = "172.25.12.0/24"
		description = "this is a fourth IPV4 address, exceeding allowed limit"
	}
`)

func consulClusterConfig(clusterID string, opt string) string {
	return fmt.Sprintf(`
	resource "hcp_consul_cluster" "test" {
		cluster_id         = "%s"
		hvn_id             = hcp_hvn.test.hvn_id
		tier               = "STANDARD"
		min_consul_version = data.hcp_consul_versions.test.recommended
	
		%s
	}
	`, clusterID, opt)
}

func setTestAccConsulClusterConfig(consulCluster string) string {
	return fmt.Sprintf(`

	resource "hcp_hvn" "test" {
		hvn_id         = "%s"
		cloud_provider = "aws"
		region         = "us-west-2"
	}

	data "hcp_consul_versions" "test" {}

	%s
	
	data "hcp_consul_cluster" "test" {
		cluster_id = hcp_consul_cluster.test.cluster_id
	}
	
	resource "hcp_consul_cluster_root_token" "test" {
		cluster_id = hcp_consul_cluster.test.cluster_id
	}
`, consulClusterHVNUniqueID, consulCluster)
}

// This includes tests against both the resource, the corresponding datasource,
// and creation of the Consul cluster root token resource in order to shorten
// testing time.
func TestAccConsulCluster(t *testing.T) {
	resourceName := "hcp_consul_cluster.test"
	dataSourceName := "data.hcp_consul_cluster.test"
	dataSourceVersionName := "data.hcp_consul_versions.test"
	rootTokenResourceName := "hcp_consul_cluster_root_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckConsulClusterDestroy,
		Steps: []resource.TestStep{
			// Tests create failure for IP Allowlist with too many CIDRs
			{
				Config:      testConfig(setTestAccConsulClusterConfig(createConsulClusterCIDRExceeded)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Too many ip_allowlist blocks`),
			},
			// Tests create
			{
				Config: testConfig(setTestAccConsulClusterConfig(consulCluster)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", consulClusterUniqueID),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", consulClusterHVNUniqueID),
					resource.TestCheckResourceAttr(resourceName, "tier", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "datacenter", consulClusterUniqueID),
					resource.TestCheckResourceAttr(resourceName, "scale", "3"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_interval", "24h"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_retention", "30d"),
					resource.TestCheckResourceAttr(resourceName, "connect_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_hvn_to_hvn_peering", "false"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_config_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_version", dataSourceVersionName, "recommended"),
					resource.TestCheckNoResourceAttr(resourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_private_endpoint_url"),
					testAccCheckFullURL(resourceName, "consul_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(resourceName, "self_link"),
					resource.TestCheckNoResourceAttr(resourceName, "primary_link"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_accessor_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_secret_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_allowlist.*", map[string]string{
						"address":     "172.25.16.0/24",
						"description": "this is an IPV4 address",
					}),
				),
			},
			// Tests import
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
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"consul_root_token_accessor_id", "consul_root_token_secret_id"},
			},
			// Tests read
			{
				Config: testConfig(setTestAccConsulClusterConfig(consulCluster)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_id", consulClusterUniqueID),
					resource.TestCheckResourceAttr(resourceName, "hvn_id", consulClusterHVNUniqueID),
					resource.TestCheckResourceAttr(resourceName, "tier", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "datacenter", consulClusterUniqueID),
					resource.TestCheckResourceAttr(resourceName, "scale", "3"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_interval", "24h"),
					resource.TestCheckResourceAttr(resourceName, "consul_snapshot_retention", "30d"),
					resource.TestCheckResourceAttr(resourceName, "connect_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_config_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_version"),
					resource.TestCheckNoResourceAttr(resourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_private_endpoint_url"),
					testAccCheckFullURL(resourceName, "consul_private_endpoint_url", ""),
					resource.TestCheckResourceAttrSet(resourceName, "self_link"),
					resource.TestCheckNoResourceAttr(resourceName, "primary_link"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_accessor_id"),
					resource.TestCheckResourceAttrSet(resourceName, "consul_root_token_secret_id"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_allowlist.*", map[string]string{
						"address":     "172.25.16.0/24",
						"description": "this is an IPV4 address",
					}),
				),
			},
			// Tests datasource
			{
				Config: testConfig(setTestAccConsulClusterConfig(consulCluster)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", dataSourceName, "project_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", dataSourceName, "organization_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hvn_id", dataSourceName, "hvn_id"),
					resource.TestCheckResourceAttrPair(resourceName, "cloud_provider", dataSourceName, "cloud_provider"),
					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "region"),
					resource.TestCheckResourceAttrPair(resourceName, "public_endpoint", dataSourceName, "public_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "datacenter", dataSourceName, "datacenter"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_enabled", dataSourceName, "connect_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_automatic_upgrades", dataSourceName, "consul_automatic_upgrades"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_snapshot_interval", dataSourceName, "consul_snapshot_interval"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_snapshot_retention", dataSourceName, "consul_snapshot_retention"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_config_file", dataSourceName, "consul_config_file"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_ca_file", dataSourceName, "consul_ca_file"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_version", dataSourceName, "consul_version"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_public_endpoint_url", dataSourceName, "consul_public_endpoint_url"),
					resource.TestCheckResourceAttrPair(resourceName, "consul_private_endpoint_url", dataSourceName, "consul_private_endpoint_url"),
					testAccCheckFullURL(resourceName, "consul_private_endpoint_url", ""),
					resource.TestCheckResourceAttrPair(resourceName, "scale", dataSourceName, "scale"),
					resource.TestCheckResourceAttrPair(resourceName, "tier", dataSourceName, "tier"),
					resource.TestCheckResourceAttrPair(resourceName, "size", dataSourceName, "size"),
					resource.TestCheckResourceAttrPair(resourceName, "self_link", dataSourceName, "self_link"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_link", dataSourceName, "primary_link"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_allowlist", dataSourceName, "ip_allowlist"),
				),
			},
			// Tests root token
			{
				Config: testConfig(setTestAccConsulClusterConfig(consulCluster)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rootTokenResourceName, "cluster_id", consulClusterUniqueID),
					resource.TestCheckResourceAttrSet(rootTokenResourceName, "accessor_id"),
					resource.TestCheckResourceAttrSet(rootTokenResourceName, "secret_id"),
					resource.TestCheckResourceAttrSet(rootTokenResourceName, "kubernetes_secret"),
				),
			},
			// Tests update
			{
				Config: testConfig(setTestAccConsulClusterConfig(updatedConsulCluster)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsulClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "size", "MEDIUM"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_allowlist.*", map[string]string{
						"address":     "172.25.14.0/24",
						"description": "this is an updated IPV4 address",
					}),
				),
			},
		},
	})
}

func testAccCheckConsulClusterExists(name string) resource.TestCheckFunc {
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

		link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
		if err != nil {
			return fmt.Errorf("unable to build link for %q: %v", id, err)
		}

		clusterID := link.ID
		loc := link.Location

		if _, err := clients.GetConsulClusterByID(context.Background(), client, loc, clusterID); err != nil {
			return fmt.Errorf("unable to read Consul cluster %q: %v", id, err)
		}

		return nil
	}
}

func testAccCheckConsulClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clients.Client)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "hcp_consul_cluster":
			id := rs.Primary.ID

			link, err := buildLinkFromURL(id, ConsulClusterResourceType, client.Config.OrganizationID)
			if err != nil {
				return fmt.Errorf("unable to build link for %q: %v", id, err)
			}

			clusterID := link.ID
			loc := link.Location

			_, err = clients.GetConsulClusterByID(context.Background(), client, loc, clusterID)
			if err == nil || !clients.IsResponseCodeNotFound(err) {
				return fmt.Errorf("didn't get a 404 when reading destroyed Consul cluster %q: %v", id, err)
			}

		default:
			continue
		}
	}
	return nil
}
