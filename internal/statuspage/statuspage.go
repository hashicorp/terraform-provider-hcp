// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statuspage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkv2Diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var (
	statuspageURL = "https://status.hashicorp.com/api/v1/summary"
	clientTimeout = 1 * time.Second
)

const (
	warnSummary   = "You may experience issues using HCP."
	warnDetailFmt = "HCP is reporting the following:\n\n%s\n\nPlease check https://status.hashicorp.com for more details."
)

// IDs from the incident.io status page for components relevant to this provider
var hcpComponentNames = map[string]string{
	"HCP API":           "01JK8R0BRY4185T4NHJFAXP35D",
	"HCP Boundary":      "01JK8R0BRYHN4JYQ1H3WC42RWV",
	"HCP Packer":        "01JK8R0BRYR9EYAGMNJ5EKC6CS",
	"HCP Portal":        "01JK8R0BRYKPJS5K35R2ZCSHV0",
	"HCP Vault Radar":   "01JK8R0BRYDYZFQH1V8ZSJKDFF",
	"HCP Vault Secrets": "01JK8R0BRYY1ZM4NCA18A5T43A",
	"HCP Waypoint":      "01JK8R0BRY0Q21819AYRKH5GZZ",
}

// These groups contain many regions and have a top-level ID that is not returned in the API response
var hcpGroupNames = []string{
	"HCP Consul Dedicated",
	"HCP Vault Dedicated",
}

type statuspage struct {
	OngoingIncidents []incident `json:"ongoing_incidents"`
}

type incident struct {
	Name               string              `json:"name"`
	Status             string              `json:"status"`
	AffectedComponents []affectedComponent `json:"affected_components"`
}

type affectedComponent struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	GroupName     string `json:"group_name,omitempty"`
	CurrentStatus string `json:"current_status"`
}

// Two types of diagnostic messages we might return
type statusCheckResult struct {
	errorMessage  string // For HTTP errors, JSON parsing errors
	statusMessage string // For actual HCP service outages
}

func (s statusCheckResult) hasDiagnostics() bool {
	return s.errorMessage != "" || s.statusMessage != ""
}

func (s statusCheckResult) diagnosticMessage() string {
	if s.errorMessage != "" {
		return s.errorMessage
	}
	if s.statusMessage != "" {
		return fmt.Sprintf(warnDetailFmt, s.statusMessage)
	}
	return ""
}

// Determine whether the components returned in the API are relevant
func isHCPComponentAffected(comp affectedComponent) bool {
	if comp.CurrentStatus == "operational" {
		return false
	}
	expectedID, ok := hcpComponentNames[comp.Name]
	if ok && expectedID == comp.ID {
		return true
	}
	return slices.Contains(hcpGroupNames, comp.GroupName)
}

// Fetch and parse the API response
func checkHCPStatus() statusCheckResult {
	var result statusCheckResult

	req, err := http.NewRequest("GET", statuspageURL, nil)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to create request to verify HCP status: %s", err)
		return result
	}

	var cl = http.Client{
		Timeout: clientTimeout,
	}
	resp, err := cl.Do(req)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to complete request to verify HCP status: %s", err)
		return result
	}
	defer resp.Body.Close()

	jsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to read response to verify HCP status: %s", err)
		return result
	}

	sp := statuspage{}
	err = json.Unmarshal(jsBytes, &sp)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to unmarshal response to verify HCP status: %s", err)
		return result
	}

	var statusBuilder strings.Builder

	for _, inc := range sp.OngoingIncidents {
		reported := make([]string, 0, len(inc.AffectedComponents))
		for _, comp := range inc.AffectedComponents {
			if isHCPComponentAffected(comp) {
				prefix := comp.Name
				if comp.GroupName != "" {
					prefix = fmt.Sprintf("%s (%s)", comp.GroupName, comp.Name)
				}
				reported = append(reported, fmt.Sprintf("%s: %s", prefix, comp.CurrentStatus))
			}
		}

		if len(reported) > 0 {
			fmt.Fprintf(&statusBuilder, "\n[status: %s] %s - %s\n",
				inc.Status, inc.Name, strings.Join(reported, ", "))
		}
	}

	result.statusMessage = statusBuilder.String()
	return result
}

func IsHCPOperationalFramework() (diags frameworkDiag.Diagnostics) {
	status := checkHCPStatus()

	if status.hasDiagnostics() {
		diags.AddWarning(warnSummary, status.diagnosticMessage())
	}

	return diags
}

func IsHCPOperationalSDKv2() (diags sdkv2Diag.Diagnostics) {
	status := checkHCPStatus()

	if status.hasDiagnostics() {
		diags = append(diags, sdkv2Diag.Diagnostic{
			Severity: sdkv2Diag.Warning,
			Summary:  warnSummary,
			Detail:   status.diagnosticMessage(),
		})
	}

	return diags
}
