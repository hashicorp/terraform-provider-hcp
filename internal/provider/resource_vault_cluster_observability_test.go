package provider

import (
	"strings"
	"testing"
)

func TestGetValidObservabilityConfig(t *testing.T) {
	cases := []struct {
		config        map[string]interface{}
		expectedError string
	}{
		{
			config: map[string]interface{}{
				"grafana_user":       "test",
				"grafana_password":   "pwd",
				"grafana_endpoint":   "https://grafana",
				"splunk_hecendpoint": "https://http-input-splunkcloud.com",
				"splunk_token":       "test",
				"datadog_api_key":    "test_datadog",
				"datadog_region":     "us1",
			},
			expectedError: "multiple configurations found: must contain configuration for only one provider",
		},
		{
			config: map[string]interface{}{
				"grafana_user":       "test",
				"grafana_password":   "",
				"grafana_endpoint":   "",
				"splunk_hecendpoint": "",
				"splunk_token":       "",
				"datadog_api_key":    "",
				"datadog_region":     "",
			},
			expectedError: "grafana configuration is invalid: configuration information missing",
		},
		{
			config: map[string]interface{}{
				"grafana_user":       "",
				"grafana_password":   "",
				"grafana_endpoint":   "",
				"splunk_hecendpoint": "",
				"splunk_token":       "test",
				"datadog_api_key":    "",
				"datadog_region":     "",
			},
			expectedError: "splunk configuration is invalid: configuration information missing",
		},
		{
			config: map[string]interface{}{
				"grafana_user":       "",
				"grafana_password":   "",
				"grafana_endpoint":   "",
				"splunk_hecendpoint": "",
				"splunk_token":       "",
				"datadog_api_key":    "",
				"datadog_region":     "us1",
			},
			expectedError: "datadog configuration is invalid: configuration information missing",
		},
	}

	for _, c := range cases {
		_, diags := getValidObservabilityConfig(c.config)
		foundError := false
		if diags.HasError() {
			for _, d := range diags {
				if strings.Contains(d.Summary, c.expectedError) {
					foundError = true
					break
				}
			}
		}
		if !foundError {
			t.Fatalf("Expected an error: %v", c.expectedError)
		}
	}
}
