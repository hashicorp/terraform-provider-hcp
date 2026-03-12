// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Vault_DisasterRecovery_PrimaryClusterAws(t *testing.T) {
	t.Parallel()

	input := &inputT{
		HvnName:          addTimestampSuffix("test-drp-hvn1-"),
		HvnCidr:          "172.21.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-primary-"),
		CloudProvider:    cloudProviderAWS,
		Region:           awsRegion,
		Tier:             "STANDARD_SMALL",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-drp-hvn2-"),
			HvnCidr:       "172.20.16.0/20",
			CloudProvider: cloudProviderAWS,
			Region:        "us-east-1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}

				data "hcp_vault_cluster" "dr" {
					cluster_id = hcp_vault_cluster.c1.cluster_id
				}
				`, input)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "cluster_id", input.VaultClusterName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "hvn_id", input.HvnName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", input.Tier),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "disaster_recovery_hvn_id", input.Secondary.HvnName),
					resource.TestCheckResourceAttr("data.hcp_vault_cluster.dr", "cluster_id", input.VaultClusterName),
					resource.TestCheckResourceAttr("data.hcp_vault_cluster.dr", "disaster_recovery_hvn_id", input.Secondary.HvnName),
				),
			},
		},
	})
}

func TestAcc_Vault_DisasterRecovery_PrimaryClusterAzure(t *testing.T) {
	t.Parallel()

	input := &inputT{
		HvnName:          addTimestampSuffix("test-drp-az-hvn1-"),
		HvnCidr:          "172.19.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-az-primary-"),
		CloudProvider:    cloudProviderAzure,
		Region:           azureRegion,
		Tier:             "STANDARD_SMALL",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-drp-az-hvn2-"),
			HvnCidr:       "172.18.16.0/20",
			CloudProvider: cloudProviderAzure,
			Region:        "eastus2",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}

				data "hcp_vault_cluster" "dr" {
					cluster_id = hcp_vault_cluster.c1.cluster_id
				}
				`, input)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultClusterExists(primaryVaultResourceName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "cluster_id", input.VaultClusterName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "hvn_id", input.HvnName),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "tier", input.Tier),
					resource.TestCheckResourceAttr(primaryVaultResourceName, "disaster_recovery_hvn_id", input.Secondary.HvnName),
					resource.TestCheckResourceAttr("data.hcp_vault_cluster.dr", "cluster_id", input.VaultClusterName),
					resource.TestCheckResourceAttr("data.hcp_vault_cluster.dr", "disaster_recovery_hvn_id", input.Secondary.HvnName),
				),
			},
		},
	})
}

func TestCIDRBlocksOverlap(t *testing.T) {
	tests := []struct {
		name    string
		cidrA   string
		cidrB   string
		overlap bool
		wantErr bool
	}{
		{
			name:    "no overlap",
			cidrA:   "10.0.0.0/16",
			cidrB:   "10.1.0.0/16",
			overlap: false,
			wantErr: false,
		},
		{
			name:    "equal CIDRs overlap",
			cidrA:   "172.25.16.0/20",
			cidrB:   "172.25.16.0/20",
			overlap: true,
			wantErr: false,
		},
		{
			name:    "one CIDR contains another",
			cidrA:   "172.25.0.0/16",
			cidrB:   "172.25.16.0/20",
			overlap: true,
			wantErr: false,
		},
		{
			name:    "adjacent CIDRs do not overlap",
			cidrA:   "172.25.16.0/20",
			cidrB:   "172.25.32.0/20",
			overlap: false,
			wantErr: false,
		},
		{
			name:    "empty CIDR invalid",
			cidrA:   "",
			cidrB:   "172.25.16.0/20",
			overlap: false,
			wantErr: true,
		},
		{
			name:    "invalid CIDR format",
			cidrA:   "invalid",
			cidrB:   "172.25.16.0/20",
			overlap: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlap, err := cidrBlocksOverlap(tt.cidrA, tt.cidrB)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if overlap != tt.overlap {
				t.Fatalf("expected overlap=%t, got overlap=%t", tt.overlap, overlap)
			}
		})
	}
}

func TestProvidersMatch(t *testing.T) {
	tests := []struct {
		name            string
		primaryProvider string
		drProvider      string
		wantMatch       bool
	}{
		{name: "same provider", primaryProvider: "aws", drProvider: "aws", wantMatch: true},
		{name: "different providers", primaryProvider: "aws", drProvider: "azure", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := providersMatch(tt.primaryProvider, tt.drProvider)
			if got != tt.wantMatch {
				t.Fatalf("expected providersMatch=%t, got=%t", tt.wantMatch, got)
			}
		})
	}
}

func TestSameRegion(t *testing.T) {
	tests := []struct {
		name            string
		primaryProvider string
		primaryRegion   string
		drProvider      string
		drRegion        string
		wantSameRegion  bool
	}{
		{
			name:            "same provider and region",
			primaryProvider: "aws",
			primaryRegion:   "us-west-2",
			drProvider:      "aws",
			drRegion:        "us-west-2",
			wantSameRegion:  true,
		},
		{
			name:            "same provider different region",
			primaryProvider: "aws",
			primaryRegion:   "us-west-2",
			drProvider:      "aws",
			drRegion:        "us-east-1",
			wantSameRegion:  false,
		},
		{
			name:            "different provider same region name",
			primaryProvider: "aws",
			primaryRegion:   "westus2",
			drProvider:      "azure",
			drRegion:        "westus2",
			wantSameRegion:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sameRegion(tt.primaryProvider, tt.primaryRegion, tt.drProvider, tt.drRegion)
			if got != tt.wantSameRegion {
				t.Fatalf("expected sameRegion=%t, got=%t", tt.wantSameRegion, got)
			}
		})
	}
}

func TestInStandardOrPlusTier(t *testing.T) {
	tests := []struct {
		name    string
		tier    string
		allowed bool
	}{
		{name: "dev tier not allowed", tier: "DEV", allowed: false},
		{name: "standard small allowed", tier: "STANDARD_SMALL", allowed: true},
		{name: "standard medium allowed", tier: "STANDARD_MEDIUM", allowed: true},
		{name: "standard large allowed", tier: "STANDARD_LARGE", allowed: true},
		{name: "plus small allowed", tier: "PLUS_SMALL", allowed: true},
		{name: "plus medium allowed", tier: "PLUS_MEDIUM", allowed: true},
		{name: "plus large allowed", tier: "PLUS_LARGE", allowed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inStandardOrPlusTier(tt.tier)
			if got != tt.allowed {
				t.Fatalf("expected inStandardOrPlusTier=%t, got=%t", tt.allowed, got)
			}
		})
	}
}

func TestAcc_Vault_DisasterRecovery_PrecheckValidationsAws(t *testing.T) {
	t.Parallel()

	// Same-region primary and DR HVNs should be rejected by provider pre-checks.
	sameRegionInput := &inputT{
		HvnName:          addTimestampSuffix("test-dr-hvn-1-"),
		HvnCidr:          "172.31.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-same-region-"),
		CloudProvider:    cloudProviderAWS,
		Region:           awsRegion,
		Tier:             "PLUS_SMALL",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-dr-hvn-2-"),
			HvnCidr:       "172.30.16.0/20",
			CloudProvider: cloudProviderAWS,
			Region:        awsRegion,
		},
	}

	// Overlapping CIDRs across different regions should be rejected by provider pre-checks.
	overlapCIDRInput := &inputT{
		HvnName:          addTimestampSuffix("test-dr-hvn-3-"),
		HvnCidr:          "172.29.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-overlap-cidr-"),
		CloudProvider:    cloudProviderAWS,
		Region:           awsRegion,
		Tier:             "PLUS_SMALL",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-dr-hvn-4-"),
			HvnCidr:       "172.29.16.0/20",
			CloudProvider: cloudProviderAWS,
			Region:        "us-east-1",
		},
	}

	// Cross-provider DR should be rejected by provider pre-checks.
	crossProviderInput := &inputT{
		HvnName:          addTimestampSuffix("test-dr-hvn-5-"),
		HvnCidr:          "172.28.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-cross-prov-"),
		CloudProvider:    cloudProviderAWS,
		Region:           awsRegion,
		Tier:             "PLUS_SMALL",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-dr-hvn-6-"),
			HvnCidr:       "172.27.16.0/20",
			CloudProvider: cloudProviderAzure,
			Region:        azureRegion,
		},
	}

	// DR on DEV tier should be rejected by provider pre-checks.
	invalidTierInput := &inputT{
		HvnName:          addTimestampSuffix("test-dr-hvn-7-"),
		HvnCidr:          "172.23.16.0/20",
		VaultClusterName: addTimestampSuffix("test-dr-invalid-tier-"),
		CloudProvider:    cloudProviderAWS,
		Region:           awsRegion,
		Tier:             "DEV",
		Secondary: &inputT{
			HvnName:       addTimestampSuffix("test-dr-hvn-8-"),
			HvnCidr:       "172.22.16.0/20",
			CloudProvider: cloudProviderAWS,
			Region:        "us-east-1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t, map[string]bool{"aws": false, "azure": false}) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVaultClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}
				`, sameRegionInput)),
				ExpectError: regexp.MustCompile(`disaster recovery HVN must be in a different region`),
			},
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}
				`, overlapCIDRInput)),
				ExpectError: regexp.MustCompile(`overlaps with disaster recovery HVN CIDR`),
			},
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}
				`, crossProviderInput)),
				ExpectError: regexp.MustCompile(`provider .* must match primary HVN provider`),
			},
			{
				Config: testConfig(setTestAccPerformanceReplicationE2E(t, `
				resource "hcp_vault_cluster" "c1" {
					cluster_id               = "{{ .VaultClusterName }}"
					hvn_id                   = hcp_hvn.hvn1.hvn_id
					tier                     = "{{ .Tier }}"
					disaster_recovery_hvn_id = hcp_hvn.hvn2.hvn_id
				}
				`, invalidTierInput)),
				ExpectError: regexp.MustCompile(`disaster recovery is supported only for STANDARD or PLUS tiers`),
			},
		},
	})
}
