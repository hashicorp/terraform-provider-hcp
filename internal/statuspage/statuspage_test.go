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

// Helper functions to create test data
func testComponent(name string, status string) affectedComponent {
	id := hcpComponentNames[name]
	return affectedComponent{
		ID:            id,
		Name:          name,
		CurrentStatus: status,
	}
}

func testGroupedComponent(groupName, status string) affectedComponent {
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
func createTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	prevURL := statuspageURL
	statuspageURL = server.URL

	t.Cleanup(func() {
		server.Close()
		statuspageURL = prevURL
	})
}

// stubStatusPage configures a test server to return a simulated status page response
func stubStatusPage(t *testing.T, incidents []incident) {
	t.Helper()
	createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(statuspage{OngoingIncidents: incidents}); err != nil {
			t.Fatalf("Failed to encode status page response: %v", err)
		}
	})
}

// Different error scenarios
func simulateError(t *testing.T, errorType string) {
	t.Helper()
	switch errorType {
	case "timeout":
		oldTimeout := clientTimeout
		t.Cleanup(func() {
			clientTimeout = oldTimeout
		})
		clientTimeout = 1 * time.Millisecond

		createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

	case "serviceDown":
		createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})
	}
}

// TestIsHCPComponentAffected tests the component evaluation function
func TestIsHCPComponentAffected(t *testing.T) {
	testCases := []struct {
		name       string
		component  affectedComponent
		isAffected bool
	}{
		{
			name:       "operational HCP component",
			component:  testComponent("HCP API", "operational"),
			isAffected: false,
		},
		{
			name:       "non-operational HCP component",
			component:  testComponent("HCP Portal", "degraded_performance"),
			isAffected: true,
		},
		{
			name:       "operational HCP group component",
			component:  testGroupedComponent("HCP Vault Dedicated", "operational"),
			isAffected: false,
		},
		{
			name:       "non-operational HCP group component",
			component:  testGroupedComponent("HCP Consul Dedicated", "partial_outage"),
			isAffected: true,
		},
		{
			name:       "non-HCP component",
			component:  testComponent("Other", "major_outage"),
			isAffected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isHCPComponentAffected(tc.component)
			assert.Equal(t, tc.isAffected, result)
		})
	}
}

// TestCheckHCPStatus tests the main status checking function
func TestCheckHCPStatus(t *testing.T) {
	testCases := []struct {
		name              string
		setup             func(t *testing.T)
		expectOutage      bool
		expectDiagnostics bool
		messageContains   []string
		messageExcludes   []string
	}{
		{
			name:              "fully operational",
			setup:             func(t *testing.T) { stubStatusPage(t, nil) },
			expectOutage:      false,
			expectDiagnostics: false,
			messageContains:   nil,
		},
		{
			name: "single non-operational HCP component",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("API Outage", "investigating",
						testComponent("HCP API", "major_outage"))})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"API Outage - HCP API: major_outage"},
		},
		{
			name: "non-operational HCP components from multiple incidents",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("API Outage", "investigating",
						testComponent("HCP API", "major_outage")),
					inc("Portal Issues", "identified",
						testComponent("HCP Portal", "partial_outage")),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"API Outage - HCP API: major_outage", "Portal Issues - HCP Portal: partial_outage"},
		},
		{
			name: "non-operational HCP componentGroup",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("Consul incident", "investigating",
						testGroupedComponent("HCP Consul Dedicated", "major_outage")),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"Consul incident - HCP Consul Dedicated (region-name): major_outage"},
		},
		{
			name: "only HCP when mixed components",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("Mixed Outage", "investigating",
						testComponent("Other", "major_outage"),
						testComponent("HCP Packer", "partial_outage")),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"Mixed Outage - HCP Packer: partial_outage"},
			messageExcludes:   []string{"Other"},
		},
		{
			name:              "service unavailable",
			setup:             func(t *testing.T) { simulateError(t, "serviceDown") },
			expectOutage:      false,
			expectDiagnostics: true,
			messageContains:   []string{"Unable to unmarshal response"},
		},
		{
			name:              "request timeout",
			setup:             func(t *testing.T) { simulateError(t, "timeout") },
			expectOutage:      false,
			expectDiagnostics: true,
			messageContains:   []string{"Unable to complete request"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			result := checkHCPStatus()

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

// Define test scenarios that can be shared between Framework and SDKv2 tests

func TestIsHCPOperational(t *testing.T) {
	scenarios := []struct {
		name            string
		setupFn         func(t *testing.T)
		expectDiags     bool
		expectedSummary string
		detailContains  []string
		detailExcludes  []string
	}{
		{
			name: "one warn overall during HCP outages",
			setupFn: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("API Outage", "investigating", testComponent("HCP API", "degraded_performance")),
					inc("Consul Issues", "identified", testGroupedComponent("HCP Consul Dedicated", "partial_outage")),
					inc("Unrelated Incident", "investigating", testComponent("Other", "major_outage")),
				})
			},
			expectDiags:     true,
			expectedSummary: warnSummary,
			detailContains:  []string{"API Outage", "Consul Issues"},
			detailExcludes:  []string{"Unrelated Incident"},
		},
		{
			name:            "setup failure warn",
			setupFn:         func(t *testing.T) { simulateError(t, "serviceDown") },
			expectDiags:     true,
			expectedSummary: warnSummary,
			detailContains:  []string{"Unable to unmarshal response"},
		},
		{
			name:        "fully operational",
			setupFn:     func(t *testing.T) { stubStatusPage(t, nil) },
			expectDiags: false,
		},
	}

	t.Run("Framework", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				scenario.setupFn(t)
				diags := IsHCPOperationalFramework()

				if !scenario.expectDiags {
					assert.Empty(t, diags, "Should have no diagnostics when operational")
					return
				}

				assert.Len(t, diags, 1, "Should have one diagnostic")
				assert.Equal(t, frameworkDiag.SeverityWarning, diags[0].Severity(), "Should use warning severity")
				assert.Equal(t, scenario.expectedSummary, diags[0].Summary(), "Should have expected summary")

				detail := diags[0].Detail()
				for _, expectedText := range scenario.detailContains {
					assert.Contains(t, detail, expectedText, "Should contain expected text")
				}
				for _, excludedText := range scenario.detailExcludes {
					assert.NotContains(t, detail, excludedText, "Should not contain excluded text")
				}
			})
		}
	})

	t.Run("SDKv2", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				scenario.setupFn(t)
				diags := IsHCPOperationalSDKv2()

				if !scenario.expectDiags {
					assert.Empty(t, diags, "Should have no diagnostics when operational")
					return
				}

				assert.Len(t, diags, 1, "Should have one diagnostic")
				assert.Equal(t, sdkv2Diag.Warning, diags[0].Severity, "Should use warning severity")
				assert.Equal(t, scenario.expectedSummary, diags[0].Summary, "Should have expected summary")

				detail := diags[0].Detail
				for _, expectedText := range scenario.detailContains {
					assert.Contains(t, detail, expectedText, "Should contain expected text")
				}
				for _, excludedText := range scenario.detailExcludes {
					assert.NotContains(t, detail, excludedText, "Should not contain excluded text")
				}
			})
		}
	})
}
