// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"strings"
	"testing"

	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
)

func TestGetValidObservabilityConfig(t *testing.T) {
	cases := map[string]struct {
		config        map[string]interface{}
		expectedError string
	}{
		"multiple providers not allowed": {
			config: map[string]interface{}{
				"grafana_user":           "test",
				"grafana_password":       "pwd",
				"grafana_endpoint":       "https://grafana",
				"splunk_hecendpoint":     "https://http-input-splunkcloud.com",
				"splunk_token":           "test",
				"datadog_api_key":        "test_datadog",
				"datadog_region":         "us1",
				"elasticsearch_user":     "test",
				"elasticsearch_password": "test_elasticsearch",
				"elasticsearch_endpoint": "https://elasticsearch",
				"newrelic_account_id":    "123456",
				"newrelic_license_key":   "abcdefg",
				"newrelic_region":        "US",
			},
			expectedError: "multiple configurations found: must contain configuration for only one provider",
		},
		"grafana missing params": {
			config: map[string]interface{}{
				"grafana_user": "test",
			},
			expectedError: "grafana configuration is invalid: configuration information missing",
		},
		"splunk missing params": {
			config: map[string]interface{}{
				"splunk_token": "test",
			},
			expectedError: "splunk configuration is invalid: configuration information missing",
		},
		"datadog missing params": {
			config: map[string]interface{}{
				"datadog_region": "us1",
			},
			expectedError: "datadog configuration is invalid: configuration information missing",
		},
		"cloudwatch missing params": {
			config: map[string]interface{}{
				"cloudwatch_access_key_id": "1111111",
			},
			expectedError: "cloudwatch configuration is invalid: configuration information missing",
		},
		"elasticsearch missing params": {
			config: map[string]interface{}{
				"elasticsearch_user": "test",
			},
			expectedError: "elasticsearch configuration is invalid: configuration information missing",
		},
		"newrelic missing params": {
			config: map[string]interface{}{
				"newrelic_account_id": "123456",
			},
			expectedError: "newrelic configuration is invalid: configuration information missing",
		},
		"http missing params": {
			config: map[string]interface{}{
				"http_uri":            "https://localhost:3000",
				"http_basic_user":     "user",
				"http_basic_password": "pass",
			},
			expectedError: "http configuration is invalid: configuration information missing",
		},
		"http invalid codec": {
			config: map[string]interface{}{
				"http_uri":    "https://localhost:3000",
				"http_method": "POST",
				"http_codec":  "SOME_VALUE",
			},
			expectedError: "http configuration is invalud: allowed values for http_codec are only \"JSON\" or \"NDJSON\"",
		},
		"http provide bearer and basic auth": {
			config: map[string]interface{}{
				"http_uri":            "https://localhost:3000",
				"http_method":         "POST",
				"http_codec":          "JSON",
				"http_basic_user":     "test",
				"http_basic_password": "pass",
				"http_bearer_token":   "111111111",
			},
			expectedError: "http configuration is invalid: either the basic or bearer authentication method can be submitted, but not both",
		},
		"http basic auth without username": {
			config: map[string]interface{}{
				"http_uri":            "https://localhost:3000",
				"http_method":         "POST",
				"http_codec":          "JSON",
				"http_basic_password": "pass",
			},
			expectedError: "http configuration is invalid: basic authentication requires username and password",
		},
		"http basic auth without password": {
			config: map[string]interface{}{
				"http_uri":        "https://localhost:3000",
				"http_method":     "POST",
				"http_codec":      "JSON",
				"http_basic_user": "test",
			},
			expectedError: "http configuration is invalid: basic authentication requires username and password",
		},
		"too many providers takes precedence over missing params": {
			config: map[string]interface{}{
				"datadog_region":           "us1",
				"cloudwatch_access_key_id": "1111111",
			},
			expectedError: "multiple configurations found: must contain configuration for only one provider",
		},
	}

	for tcName, c := range cases {
		t.Run(tcName, func(t *testing.T) {
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
		})
	}
}

func TestGetValidMajorVersionUpgradeConfig(t *testing.T) {
	cases := []struct {
		config        map[string]interface{}
		expectedError string
		tier          vaultmodels.HashicorpCloudVault20201125Tier
	}{
		{
			config: map[string]interface{}{
				"upgrade_type":            "MANUAL",
				"maintenance_window_day":  "",
				"maintenance_window_time": "",
			},
			expectedError: "major version configuration is only allowed for STANDARD or PLUS clusters",
			tier:          vaultmodels.HashicorpCloudVault20201125TierSTARTERSMALL,
		},
		{
			config: map[string]interface{}{
				"upgrade_type":            "AUTOMATIC",
				"maintenance_window_day":  "",
				"maintenance_window_time": "",
			},
			expectedError: "major version configuration is only allowed for STANDARD or PLUS clusters",
			tier:          vaultmodels.HashicorpCloudVault20201125TierDEV,
		},

		{
			config: map[string]interface{}{
				"upgrade_type":            "MANUAL",
				"maintenance_window_day":  "SUNDAY",
				"maintenance_window_time": "WINDOW_6PM_10PM",
			},
			expectedError: "major version upgrade configuration is invalid: maintenance window is only allowed to SCHEDULED upgrades",
			tier:          vaultmodels.HashicorpCloudVault20201125TierPLUSMEDIUM,
		},
		{
			config: map[string]interface{}{
				"upgrade_type":            "AUTOMATIC",
				"maintenance_window_day":  "WEDNESDAY",
				"maintenance_window_time": "WINDOW_6AM_10AM",
			},
			expectedError: "major version upgrade configuration is invalid: maintenance window is only allowed to SCHEDULED upgrades",
			tier:          vaultmodels.HashicorpCloudVault20201125TierPLUSSMALL,
		},
		{
			config: map[string]interface{}{
				"upgrade_type":            "SCHEDULED",
				"maintenance_window_day":  "THURSDAY",
				"maintenance_window_time": "",
			},
			expectedError: "major version upgrade configuration is invalid: maintenance window configuration information missing",
			tier:          vaultmodels.HashicorpCloudVault20201125TierSTANDARDLARGE,
		},
		{
			config: map[string]interface{}{
				"upgrade_type":            "SCHEDULED",
				"maintenance_window_day":  "",
				"maintenance_window_time": "WINDOW_12PM_4PM",
			},
			expectedError: "major version upgrade configuration is invalid: maintenance window configuration information missing",
			tier:          vaultmodels.HashicorpCloudVault20201125TierSTANDARDMEDIUM,
		},
		{
			config: map[string]interface{}{
				"upgrade_type":            "SCHEDULED",
				"maintenance_window_day":  "",
				"maintenance_window_time": "",
			},
			expectedError: "major version upgrade configuration is invalid: maintenance window configuration information missing",
			tier:          vaultmodels.HashicorpCloudVault20201125TierSTANDARDSMALL,
		},
	}

	for _, c := range cases {
		_, diags := getValidMajorVersionUpgradeConfig(c.config, c.tier)
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
