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

const (
	statuspageURL = "https://status.hashicorp.com/api/v1/summary"
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

type statusCheckResult struct {
	errorMessage  string
	operational   bool
	statusMessage string
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

func checkHCPStatus() statusCheckResult {
	result := statusCheckResult{
		operational: true,
	}

	req, err := http.NewRequest("GET", statuspageURL, nil)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to create request to verify HCP status: %s", err)
		return result
	}

	var cl = http.Client{
		Timeout: 3 * time.Second,
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
				result.operational = false
				prefix := comp.Name
				if comp.GroupName != "" {
					prefix = fmt.Sprintf("%s (%s)", comp.GroupName, comp.Name)
				}
				reported = append(reported, fmt.Sprintf("%s: %s", prefix, comp.CurrentStatus))
			}
		}

		if len(reported) > 0 {
			fmt.Fprintf(&statusBuilder, "\n[Status: %s] %s\n- %s\n",
				inc.Status, inc.Name, strings.Join(reported, ", "))
		}
	}

	result.statusMessage = statusBuilder.String()
	return result
}

func IsHCPOperationalFramework() (diags frameworkDiag.Diagnostics) {

	result := checkHCPStatus()

	if result.errorMessage != "" {
		diags.AddWarning(warnSummary, result.errorMessage)
		return diags
	}

	if !result.operational {
		diags.AddWarning(warnSummary, fmt.Sprintf(warnDetailFmt, result.statusMessage))
	}

	return diags
}

func IsHCPOperationalSDKv2() (diags sdkv2Diag.Diagnostics) {

	result := checkHCPStatus()

	if result.errorMessage != "" {
		diags = append(diags, sdkv2Diag.Diagnostic{
			Severity: sdkv2Diag.Warning,
			Summary:  warnSummary,
			Detail:   result.errorMessage,
		})
		return diags
	}

	if !result.operational {
		diags = append(diags, sdkv2Diag.Diagnostic{
			Severity: sdkv2Diag.Warning,
			Summary:  warnSummary,
			Detail:   fmt.Sprintf(warnDetailFmt, result.statusMessage),
		})
	}

	return diags
}
