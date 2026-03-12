// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import "testing"

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
