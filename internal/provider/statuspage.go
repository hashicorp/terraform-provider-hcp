// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "time"

        diagnostic "github.com/hashicorp/terraform-plugin-framework/diag"
)

// Status endpoint for prod.
const statuspageURL = "https://status.hashicorp.com/api/v1/summary"

var hcpComponentNames = map[string]string{
        "HCP API":      "01JK8R0BRY4185T4NHJFAXP35D",
        "HCP Consul":   "01JKJP1CYEVMQKJ1Z5WAX0W12F", // Mapped to the HCP Consul Dedicated group
        "HCP Packer":   "01JK8R0BRYR9EYAGMNJ5EKC6CS",
        "HCP Portal":   "01JK8R0BRYKPJS5K35R2ZCSHV0",
        "HCP Vault":    "01JKJP1CYEVMQKJ1Z5WESBEQ7M", // Mapped to the HCP Vault Dedicated group
        "HCP Boundary": "01JK8R0BRYHN4JYQ1H3WC42RWV",
}

type statuspage struct {
        OngoingIncidents []incident `json:"ongoing_incidents"`
}

type incident struct {
        Name             string          `json:"name"`
        Status           string          `json:"status"`
        CurrentWorstImpact string          `json:"current_worst_impact"`
        AffectedComponents []affectedComponent `json:"affected_components"`
        LastUpdateAt     string          `json:"last_update_at"`
        LastUpdateMessage string          `json:"last_update_message"`
}

type affectedComponent struct {
        ID        string `json:"id"`
        Name      string `json:"name"`
        GroupName string `json:"group_name,omitempty"`
}

func isHCPOperationalFramework() (diags diagnostic.Diagnostics) {
        req, err := http.NewRequest("GET", statuspageURL, nil)
        if err != nil {
                diags.AddWarning("You may experience issues using HCP.",
                        fmt.Sprintf("Unable to create request to verify HCP status: %s", err))
                return diags
        }

        var cl = http.Client{}
        resp, err := cl.Do(req)
        if err != nil {
                diags.AddWarning("You may experience issues using HCP.",
                        fmt.Sprintf("Unable to create request to verify HCP status: %s", err))
                return diags
        }
        defer resp.Body.Close()

        jsBytes, err := io.ReadAll(resp.Body)
        if err != nil {
                diags.AddWarning("You may experience issues using HCP.",
                        fmt.Sprintf("Unable read response to verify HCP status: %s", err))
                return diags
        }

    	sp := statuspage{}
    	err = json.Unmarshal(jsBytes, &sp)
        if err != nil {
                diags.AddWarning("You may experience issues using HCP.",
                        fmt.Sprintf("Unable unmarshal response to verify HCP status: %s", err))
                return diags
        }
	
    	systemStatus := make(map[string]string)
    	operational := true
    	statusMessage := ""
	
    	for _, inc := range sp.OngoingIncidents {
    	    for _, comp := range inc.AffectedComponents {
    	        if expectedID, ok := hcpComponentNames[comp.Name]; ok && expectedID == comp.ID {
    	            impact := inc.CurrentWorstImpact
    	            systemStatus[comp.Name] = impact
    	            if impact != "operational" {
    	                operational = false
    	                statusMessage += fmt.Sprintf("%s: %s (%s - %s)\n", comp.Name, impact, inc.Name, inc.LastUpdateMessage)
    	            }
    	        }
    	    }
    	}
	
    	if !operational {
    	    diags.AddWarning("You may experience issues using HCP.",
    	        fmt.Sprintf("HCP is reporting the following:\n\n%s\nPlease check https://status.hashicorp.com for more details.", statusMessage))
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
