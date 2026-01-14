// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statuspage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkv2Diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

// copy of usConfig as of Jan 14 2026
var testRegionalConfig = regionalConfig{
	componentNames: map[string]string{
		"HCP API":           "01JK8R0BRY4185T4NHJFAXP35D",
		"HCP Boundary":      "01JK8R0BRYHN4JYQ1H3WC42RWV",
		"HCP Packer":        "01JK8R0BRYR9EYAGMNJ5EKC6CS",
		"HCP Portal":        "01JK8R0BRYKPJS5K35R2ZCSHV0",
		"HCP Vault Radar":   "01JK8R0BRYDYZFQH1V8ZSJKDFF",
		"HCP Vault Secrets": "01JK8R0BRYY1ZM4NCA18A5T43A",
		"HCP Waypoint":      "01JK8R0BRY0Q21819AYRKH5GZZ",
	},
	groupNames: []string{
		"HCP Consul Dedicated",
		"HCP Vault Dedicated",
		"API",
	},
	statusPageURL: "https://status.hashicorp.com/api/v1/summary",
	clientTimeout: 1,
}

// Helper functions to create test data
func testComponent(name string, status string, region *regionalConfig) affectedComponent {
	id := region.componentNames[name]

	return affectedComponent{
		ID:            id,
		Name:          name,
		CurrentStatus: status,
	}
}

func testGroupedComponent(groupName, status string, region *regionalConfig) affectedComponent {
	return affectedComponent{
		ID:            "region-id",
		Name:          "region-name",
		GroupName:     groupName,
		CurrentStatus: status,
	}
}

func inc(name, status string, components ...affectedComponent) incident {
	return incident{
		Name:               name,
		Status:             status,
		AffectedComponents: components,
	}
}

// createTestServer creates and manages a test HTTP server
func createTestServer(t *testing.T, region *regionalConfig, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	prevURL := region.statusPageURL
	region.statusPageURL = server.URL

	t.Cleanup(func() {
		server.Close()
		region.statusPageURL = prevURL
	})
}

// stubStatusPage configures a test server to return a simulated status page response
func stubStatusPage(t *testing.T, region *regionalConfig, incidents []incident) {
	t.Helper()
	createTestServer(t, region, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(statuspage{OngoingIncidents: incidents}); err != nil {
			t.Fatalf("Failed to encode status page response: %v", err)
		}
	})
}

// Different error scenarios
func simulateError(t *testing.T, errorType string, region *regionalConfig) {
	t.Helper()
	switch errorType {
	case "timeout":
		oldTimeout := region.clientTimeout
		t.Cleanup(func() {
			region.clientTimeout = oldTimeout
		})
		region.clientTimeout = 1

		createTestServer(t, region, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

	case "serviceDown":
		createTestServer(t, region, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})
	}
}

func TestIsHCPComponentAffected(t *testing.T) {

	testCases := []struct {
		name           string
		component      affectedComponent
		isAffected     bool
		regionalConfig *regionalConfig
	}{
		{
			name:           "operational HCP component",
			component:      testComponent("HCP API", "operational", &testRegionalConfig),
			isAffected:     false,
			regionalConfig: &testRegionalConfig,
		},
		{
			name:           "non-operational HCP component",
			component:      testComponent("HCP Portal", "degraded_performance", &testRegionalConfig),
			isAffected:     true,
			regionalConfig: &testRegionalConfig,
		},
		{
			name:           "operational HCP group component",
			component:      testGroupedComponent("HCP Vault Dedicated", "operational", &testRegionalConfig),
			isAffected:     false,
			regionalConfig: &testRegionalConfig,
		},
		{
			name:           "non-operational HCP group component",
			component:      testGroupedComponent("HCP Consul Dedicated", "partial_outage", &testRegionalConfig),
			isAffected:     true,
			regionalConfig: &testRegionalConfig,
		},
		{
			name:           "non-HCP component",
			component:      testComponent("Other", "major_outage", &testRegionalConfig),
			isAffected:     false,
			regionalConfig: &testRegionalConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isHCPComponentAffected(tc.component, tc.regionalConfig)
			assert.Equal(t, tc.isAffected, result)
		})
	}
}

func TestCheckHCPStatus(t *testing.T) {
	testCases := []struct {
		name              string
		setup             func(t *testing.T)
		expectOutage      bool
		expectDiagnostics bool
		messageContains   []string
		messageExcludes   []string
		geography         string
	}{
		{
			name:              "fully operational",
			setup:             func(t *testing.T) { stubStatusPage(t, &testRegionalConfig, nil) },
			expectOutage:      false,
			expectDiagnostics: false,
			messageContains:   nil,
			geography:         "us",
		},
		{
			name: "resolved incident",
			setup: func(t *testing.T) {
				stubStatusPage(t, &testRegionalConfig, []incident{
					inc("Packer issues", "monitoring", testComponent("HCP Packer", "operational", &testRegionalConfig)),
				})
			},
			expectOutage:      false,
			expectDiagnostics: false,
			messageContains:   nil,
			geography:         "us",
		},
		{
			name: "multi-component incident",
			setup: func(t *testing.T) {
				stubStatusPage(t, &testRegionalConfig, []incident{
					inc("Mixed issues", "investigating",
						testComponent("HCP Boundary", "degraded_performance", &testRegionalConfig),
						testGroupedComponent("HCP Consul Dedicated", "degraded_performance", &testRegionalConfig),
						testComponent("HCP Waypoint", "operational", &testRegionalConfig),
						testComponent("Other Service", "major_outage", &testRegionalConfig)),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"HCP Boundary", "HCP Consul Dedicated"},
			messageExcludes:   []string{"HCP Waypoint", "Other Service"},
			geography:         "us",
		},
		{
			name: "multiple incidents",
			setup: func(t *testing.T) {
				stubStatusPage(t, &testRegionalConfig, []incident{
					inc("HCP Vault Radar", "identified", testComponent("HCP Vault Radar", "partial_outage", &testRegionalConfig)),
					inc("HCP Vault Secrets", "investigating", testComponent("HCP Vault Secrets", "major_outage", &testRegionalConfig)),
					inc("HCP Vault Dedicated", "investigating", testGroupedComponent("HCP Vault Dedicated", "partial_outage", &testRegionalConfig)),
					inc("Other Service", "investigating", testComponent("Other Service", "major_outage", &testRegionalConfig)),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"HCP Vault Radar", "HCP Vault Secrets", "HCP Vault Dedicated (region-name)"},
			messageExcludes:   []string{"Other Service"},
			geography:         "us",
		},
		{
			name:              "service unavailable",
			setup:             func(t *testing.T) { simulateError(t, "serviceDown", &testRegionalConfig) },
			expectOutage:      false,
			expectDiagnostics: true,
			messageContains:   []string{"Unable to unmarshal response"},
			geography:         "us",
		},
		{
			name:              "request timeout",
			setup:             func(t *testing.T) { simulateError(t, "timeout", &testRegionalConfig) },
			expectOutage:      false,
			expectDiagnostics: true,
			messageContains:   []string{"Unable to complete request"},
			geography:         "us",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			result := checkHCPStatus(&tc.geography)

			assert.Equal(t, tc.expectOutage, result.statusMessage != "", "Operational status mismatch")
			assert.Equal(t, tc.expectDiagnostics, result.hasDiagnostics(), "Diagnostics presence mismatch")

			if tc.expectDiagnostics {
				message := result.diagnosticMessage()
				assert.NotEmpty(t, message, "Expected diagnostic message to be non-empty")

				for _, s := range tc.messageContains {
					assert.Contains(t, message, s, "Diagnostic message should contain '%s'", s)
				}
				for _, s := range tc.messageExcludes {
					assert.NotContains(t, message, s, "Diagnostic message should not contain '%s'", s)
				}
			} else {
				assert.Empty(t, result.diagnosticMessage(), "Expected no diagnostic message")
			}
		})
	}
}

func TestIsHCPOperational(t *testing.T) {
	scenarios := []struct {
		name            string
		setupFn         func(t *testing.T)
		expectDiags     bool
		expectedSummary string
		detailContains  []string
		detailExcludes  []string
		geography       string
	}{
		{
			name: "one warn overall during HCP outages",
			setupFn: func(t *testing.T) {
				stubStatusPage(t, &testRegionalConfig, []incident{
					inc("API Outage", "investigating", testComponent("HCP API", "degraded_performance", &testRegionalConfig)),
					inc("Consul Issues", "identified", testGroupedComponent("HCP Consul Dedicated", "partial_outage", &testRegionalConfig)),
					inc("Unrelated Incident", "investigating", testComponent("Other", "major_outage", &testRegionalConfig)),
				})
			},
			expectDiags:     true,
			expectedSummary: warnSummary,
			detailContains:  []string{"API Outage", "Consul Issues"},
			detailExcludes:  []string{"Unrelated Incident"},
			geography:       "us",
		},
		{
			name:            "setup failure warn",
			setupFn:         func(t *testing.T) { simulateError(t, "serviceDown", &testRegionalConfig) },
			expectDiags:     true,
			expectedSummary: warnSummary,
			detailContains:  []string{"Unable to unmarshal response"},
			detailExcludes:  []string{"HCP is reporting the following"},
			geography:       "us",
		},
		{
			name:        "fully operational",
			setupFn:     func(t *testing.T) { stubStatusPage(t, &testRegionalConfig, nil) },
			expectDiags: false,
			geography:   "us",
		},
	}

	implementations := []string{"Framework", "SDKv2"}

	for _, impl := range implementations {
		for _, scenario := range scenarios {
			t.Run(impl+" "+scenario.name, func(t *testing.T) {
				scenario.setupFn(t)

				var diags interface{}
				if impl == "Framework" {
					diags = IsHCPOperationalFramework(scenario.geography)
				}
				if impl == "SDKv2" {
					diags = IsHCPOperationalSDKv2(scenario.geography)
				}

				if !scenario.expectDiags {
					assert.Empty(t, diags, "Should have no diagnostics when operational")
					return
				}

				assert.Len(t, diags, 1, "Should have one diagnostic")

				switch d := diags.(type) {
				case frameworkDiag.Diagnostics:
					assert.Equal(t, frameworkDiag.SeverityWarning, d[0].Severity(), "Severity should be Warning not Error")
					assert.Equal(t, scenario.expectedSummary, d[0].Summary())

					for _, expectedText := range scenario.detailContains {
						assert.Contains(t, d[0].Detail(), expectedText)
					}
					for _, excludedText := range scenario.detailExcludes {
						assert.NotContains(t, d[0].Detail(), excludedText)
					}
				case sdkv2Diag.Diagnostics:
					assert.Equal(t, sdkv2Diag.Warning, d[0].Severity, "Severity should be Warning not Error")
					assert.Equal(t, scenario.expectedSummary, d[0].Summary)

					for _, expectedText := range scenario.detailContains {
						assert.Contains(t, d[0].Detail, expectedText)
					}
					for _, excludedText := range scenario.detailExcludes {
						assert.NotContains(t, d[0].Detail, excludedText)
					}
				}
			})
		}
	}
}
