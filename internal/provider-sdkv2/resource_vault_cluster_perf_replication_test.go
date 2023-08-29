// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	hvn1ResourceName           = "hcp_hvn.hvn1"
	hvn2ResourceName           = "hcp_hvn.hvn2"
	primaryVaultResourceName   = "hcp_vault_cluster.c1"
	secondaryVaultResourceName = "hcp_vault_cluster.c2"
)

func setTestAccPerformanceReplicationE2E(t *testing.T, tfCode string, in *inputT) string {
	templates := fmt.Sprintf(`
	resource "hcp_hvn" "hvn1" {
		hvn_id            = "{{ .HvnName }}"
		cidr_block        = "{{ .GetHvnCidr }}"
		cloud_provider    = "{{ .CloudProvider }}"
		region            = "{{ .Region }}"
	}
	
	resource "hcp_hvn" "hvn2" {
		hvn_id            = "{{ .Secondary.HvnName }}"
		cidr_block        = "{{ .Secondary.GetHvnCidr }}"
		cloud_provider    = "{{ .Secondary.CloudProvider }}"
		region            = "{{ .Secondary.Region }}"
	}
	
	%s
	`, tfCode)

	tmpl, err := template.New("tf_resources_perf_repl").Parse(templates)
	require.NoError(t, err)

	tfResources := &bytes.Buffer{}
	err = tmpl.Execute(tfResources, in)
	require.NoError(t, err)
	return tfResources.String()
}

func TestAccPerformanceReplication_ValidationsAws(t *testing.T) {
	awsPerfReplicationTestInput := &inputT{
		HvnName:                  addTimestampSuffix("test-perf-hvn-1-"),
		HvnCidr:                  "172.25.16.0/20",
		VaultClusterName:         addTimestampSuffix("test-primary-aws-"),
		VaultClusterResourceName: primaryVaultResourceName,
		CloudProvider:            cloudProviderAWS,
		Region:                   awsRegion,
		Tier:                     "PLUS_SMALL",
		UpdateTier1:              "PLUS_MEDIUM",
		UpdateTier2:              "STARTER_SMALL",
		Secondary: &inputT{
			HvnName:                  addTimestampSuffix("test-perf-hvn-2-"),
			HvnCidr:                  "172.24.16.0/20",
			VaultClusterName:         addTimestampSuffix("test-secondary-aws-"),
			VaultClusterResourceName: secondaryVaultResourceName,
			CloudProvider:            cloudProviderAWS,
			Region:                   awsRegion,
			Tier:                     "PLUS_SMALL",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckVaultClusterDestroy,
		Steps:             performanceReplicationSteps(t, awsPerfReplicationTestInput),
	})
}

func performanceReplicationSteps(t *testing.T, in *inputT) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, "", in)),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(hvn1ResourceName, "hvn_id", in.HvnName),
				resource.TestCheckResourceAttr(hvn1ResourceName, "cidr_block", in.HvnCidr),
				resource.TestCheckResourceAttr(hvn2ResourceName, "hvn_id", in.Secondary.HvnName),
				resource.TestCheckResourceAttr(hvn2ResourceName, "cidr_block", in.Secondary.HvnCidr),
			),
		},
		{
			// invalid primary link supplied
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = "{{ .Secondary.Tier }}"
				primary_link = "something"
				public_endpoint = true
			}
			`, in)),
			ExpectError: regexp.MustCompile(`invalid primary_link supplied*`),
		},
		{
			// incorrectly specify a paths_filter on a non-secondary
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id   = "{{ .VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = "{{ .Tier }}"
				paths_filter = ["path/a"]
			}
			`, in)),
			ExpectError: regexp.MustCompile(`only performance replication secondaries may specify a paths_filter`),
		},
		{
			// create a plus tier cluster successfully
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckVaultClusterExists(primaryVaultResourceName),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "cluster_id", in.VaultClusterName),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "hvn_id", in.HvnName),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", in.Tier),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "cloud_provider", in.CloudProvider),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "region", in.Region),
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
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = "plus_medium"
				primary_link = hcp_vault_cluster.c1.self_link
			}
			`, in)),
			ExpectError: regexp.MustCompile(`a secondary's tier must match that of its primary`),
		},
		{
			// secondary cluster creation failed as primary link is invalid
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = hcp_vault_cluster.c1.tier
				primary_link = "not-present"
			}
			`, in)),
			ExpectError: regexp.MustCompile(`invalid primary_link supplied url`),
		},
		{
			// secondary cluster creation failed as min_vault_version is specified.
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id        = "{{ .Secondary.VaultClusterName }}"
				hvn_id            = hcp_hvn.hvn1.hvn_id
				tier              = hcp_vault_cluster.c1.tier
				primary_link      = hcp_vault_cluster.c1.self_link
				min_vault_version = "v1.0.1"
			}
			`, in)),
			ExpectError: regexp.MustCompile(`min_vault_version should either be unset or match the primary cluster's`),
		},
		{
			// secondary cluster created successfully (same hvn)
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = lower(hcp_vault_cluster.c1.tier)
				primary_link = hcp_vault_cluster.c1.self_link
				paths_filter = ["path/a", "path/b"]
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckVaultClusterExists(primaryVaultResourceName),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "cluster_id", in.Secondary.VaultClusterName),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "hvn_id", in.HvnName), // same as primary
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", in.Secondary.Tier),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "cloud_provider", in.CloudProvider), // same as primary
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "region", in.Region),                // samae as primary
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
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = hcp_vault_cluster.c1.tier
				primary_link = hcp_vault_cluster.c1.self_link
				paths_filter = ["path/a", "path/c"]
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.0", "path/a"),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "paths_filter.1", "path/c"),
			),
		},
		{
			// delete paths filter
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn1.hvn_id
				tier         = hcp_vault_cluster.c1.tier
				primary_link = hcp_vault_cluster.c1.self_link
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr(secondaryVaultResourceName, "paths_filter.0"),
			),
		},
		{
			// secondary cluster created successfully (different hvn)
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .Tier }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn2.hvn_id
				tier         = lower(hcp_vault_cluster.c1.tier)
				primary_link = hcp_vault_cluster.c1.self_link
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckVaultClusterExists(primaryVaultResourceName),
				testAccCheckVaultClusterExists(secondaryVaultResourceName),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "cluster_id", in.Secondary.VaultClusterName),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "hvn_id", in.Secondary.HvnName),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", in.Secondary.Tier),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "cloud_provider", in.Secondary.CloudProvider),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "region", in.Secondary.Region),
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
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .UpdateTier1 }}"
				public_endpoint = true
			}
			resource "hcp_vault_cluster" "c2" {
				cluster_id   = "{{ .Secondary.VaultClusterName }}"
				hvn_id       = hcp_hvn.hvn2.hvn_id
				tier         = lower(hcp_vault_cluster.c1.tier)
				primary_link = hcp_vault_cluster.c1.self_link
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", in.UpdateTier1),
				resource.TestCheckResourceAttr(secondaryVaultResourceName, "tier", in.UpdateTier1),
			),
		},
		{
			// successfully disable replication
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .UpdateTier1 }}"
				public_endpoint = true
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckVaultClusterExists(primaryVaultResourceName),
			),
		},
		{
			// successfully scale out of the Plus tier
			Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
			resource "hcp_vault_cluster" "c1" {
				cluster_id      = "{{ .VaultClusterName }}"
				hvn_id          = hcp_hvn.hvn1.hvn_id
				tier            = "{{ .UpdateTier2 }}"
				public_endpoint = true
			}
			`, in)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckVaultClusterExists(primaryVaultResourceName),
				resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", in.UpdateTier2),
			),
		},
	}
}
