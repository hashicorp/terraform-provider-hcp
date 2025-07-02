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
			name: "resolved incident",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("Packer issues", "monitoring", testComponent("HCP Packer", "operational")),
				})
			},
			expectOutage:      false,
			expectDiagnostics: false,
			messageContains:   nil,
		},
		{
			name: "multi-component incident",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("Mixed issues", "investigating",
						testComponent("HCP Boundary", "degraded_performance"),
						testGroupedComponent("HCP Consul Dedicated", "degraded_performance"),
						testComponent("HCP Waypoint", "operational"),
						testComponent("Other Service", "major_outage")),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"HCP Boundary", "HCP Consul Dedicated"},
			messageExcludes:   []string{"HCP Waypoint", "Other Service"},
		},
		{
			name: "multiple incidents",
			setup: func(t *testing.T) {
				stubStatusPage(t, []incident{
					inc("HCP Vault Radar", "identified", testComponent("HCP Vault Radar", "partial_outage")),
					inc("HCP Vault Secrets", "investigating", testComponent("HCP Vault Secrets", "major_outage")),
					inc("HCP Vault Dedicated", "investigating", testGroupedComponent("HCP Vault Dedicated", "partial_outage")),
					inc("Other Service", "investigating", testComponent("Other Service", "major_outage")),
				})
			},
			expectOutage:      true,
			expectDiagnostics: true,
			messageContains:   []string{"HCP Vault Radar", "HCP Vault Secrets", "HCP Vault Dedicated (region-name)"},
			messageExcludes:   []string{"Other Service"},
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
			detailExcludes:  []string{"HCP is reporting the following"},
		},
		{
			name:        "fully operational",
			setupFn:     func(t *testing.T) { stubStatusPage(t, nil) },
			expectDiags: false,
		},
	}

	implementations := []string{"Framework", "SDKv2"}

	for _, impl := range implementations {
		for _, scenario := range scenarios {
			t.Run(impl+" "+scenario.name, func(t *testing.T) {
				scenario.setupFn(t)

				var diags interface{}
				if impl == "Framework" {
					diags = IsHCPOperationalFramework()
				}
				if impl == "SDKv2" {
					diags = IsHCPOperationalSDKv2()
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
