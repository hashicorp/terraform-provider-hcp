// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Status endpoint for prod.
const statuspageURL = "https://status.hashicorp.com/api/v2/components.json"

var hcpComponentIds = map[string]string{
	"0q55nwmxngkc": "HCP API",
	"sxffkgfb4fhb": "HCP Consul",
	"0mbkqnrzg33w": "HCP Packer",
	"mgv1p2j9x444": "HCP Portal",
	"mb7xrbx9gjnq": "HCP Vault",
}

type statuspage struct {
	Components []component `json:"components"`
}

type component struct {
	ID     string `json:"id"`
	Status status `json:"status"`
}

type status string

func isHCPOperational() (diags diag.Diagnostics) {
	req, err := http.NewRequest("GET", statuspageURL, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You may experience issues using HCP.",
			Detail:   fmt.Sprintf("Unable to create request to verify HCP status: %s", err),
		})

		return diags
	}

	var cl = http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You may experience issues using HCP.",
			Detail:   fmt.Sprintf("Unable to complete request to verify HCP status: %s", err),
		})

		return diags
	}
	defer resp.Body.Close()

	jsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You may experience issues using HCP.",
			Detail:   fmt.Sprintf("Unable read response to verify HCP status: %s", err),
		})

		return diags
	}

	sp := statuspage{}
	err = json.Unmarshal(jsBytes, &sp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You may experience issues using HCP.",
			Detail:   fmt.Sprintf("Unable unmarshal response to verify HCP status: %s", err),
		})

		return diags
	}

	// Translate the status page component IDs into a map of component name and operation status.
	var systemStatus = map[string]status{}

	for _, c := range sp.Components {
		name, ok := hcpComponentIds[c.ID]
		if ok {
			systemStatus[name] = c.Status
		}
	}

	operational := true
	for _, st := range systemStatus {
		if st != "operational" {
			operational = false
		}
	}

	if !operational {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You may experience issues using HCP.",
			Detail:   fmt.Sprintf("HCP is reporting the following:\n\n%v\nPlease check https://status.hashicorp.com for more details.", printStatus(systemStatus)),
		})
	}

	return diags
}

func printStatus(m map[string]status) string {
	var maxLenKey int
	for k := range m {
		if len(k) > maxLenKey {
			maxLenKey = len(k)
		}
	}

	pr := ""
	for k, v := range m {
		pr += fmt.Sprintf("%s:%*s %s\n", k, 5+(maxLenKey-len(k)), " ", v)
	}

	return pr
}
