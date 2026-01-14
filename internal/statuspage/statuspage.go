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
	warnSummary   = "You may experience issues using HCP."
	warnDetailFmt = "HCP is reporting the following:\n\n%s\n\nPlease check https://status.hashicorp.com for more details."
)

type regionalConfig struct {
	// Creating components on the incident.io status page generates a unique ID that can be found in the DOM and API response
	// Component names are not unique, so we include both here to ensure we report on the correct components
	componentNames map[string]string

	// Components can be grouped into named folders, in this case containing many cloud regions
	// Groups have a top-level ID that are not returned in the API response so we rely on the group names
	groupNames []string

	statusPageURL string
	name          string
	clientTimeout int
}

// Creating components on the incident.io status page generates a unique ID that can be found in the DOM and API response
// Component names are not unique, so we include both here to ensure we report on the correct components
var euConfig = regionalConfig{
	componentNames: map[string]string{
		"HCP API":       "01K7FC148SJVJT1TNY9VG83DTE",
		"HCP Portal":    "01K7FC148SHCEJVZZXH0DPNMC2",
		"HCP Terraform": "01K7FC148SPG6CET2XAH6GFCC7",
		"HCP Waypoint":  "01K7FC148SGM8V154MQ73CWVF6",
		"Portal":        "01JADGGSJTM1102ZE8F65Q3F56",
	},
	groupNames: []string{
		"HCP Consul Dedicated",
		"HCP Vault Dedicated",
		"API",
	},
	statusPageURL: "https://status.eu.hashicorp.com/api/v1/summary",
	clientTimeout: 1,
}

var sandboxConfig = regionalConfig{
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
	statusPageURL: "https://statuspage.incident.io/status-page-sandbox-8162182495/api/v1/summary",
	clientTimeout: 1,
}

var usConfig = regionalConfig{
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

var regions = map[string]regionalConfig{
	"eu":      euConfig,
	"sandbox": sandboxConfig,
	"us":      usConfig,
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
func isHCPComponentAffected(comp affectedComponent, region *regionalConfig) bool {
	if comp.CurrentStatus == "operational" {
		return false
	}
	expectedID, ok := region.componentNames[comp.Name]
	if ok && expectedID == comp.ID {
		return true
	}
	return slices.Contains(region.groupNames, comp.GroupName)
}

// Fetch and parse the API response
func checkHCPStatus(geography *string) statusCheckResult {
	var result statusCheckResult
	var statusBuilder strings.Builder
	var reported []string

	region, oK := regions[*geography]
	if !oK {
		region = regions["us"]
	}

	fmt.Printf("Region name is %s", region)
	statuspageURL := region.statusPageURL

	req, err := http.NewRequest("GET", statuspageURL, nil)
	if err != nil {
		result.errorMessage = fmt.Sprintf("Unable to create request to verify HCP status: %s", err)
		return result
	}

	cl := &http.Client{
		Timeout: time.Duration(region.clientTimeout) * time.Second,
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

	for _, inc := range sp.OngoingIncidents {
		for _, comp := range inc.AffectedComponents {
			if isHCPComponentAffected(comp, &region) {
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

func IsHCPOperationalFramework(geography string) (diags frameworkDiag.Diagnostics) {
	status := checkHCPStatus(&geography)

	if status.hasDiagnostics() {
		diags.AddWarning(warnSummary, status.diagnosticMessage())
	}

	return diags
}

func IsHCPOperationalSDKv2(geography string) (diags sdkv2Diag.Diagnostics) {
	status := checkHCPStatus(&geography)

	if status.hasDiagnostics() {
		diags = append(diags, sdkv2Diag.Diagnostic{
			Severity: sdkv2Diag.Warning,
			Summary:  warnSummary,
			Detail:   status.diagnosticMessage(),
		})
	}

	return diags
}
