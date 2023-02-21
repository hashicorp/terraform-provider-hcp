// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

const (
	cloudProviderAWS           = "aws"
	cloudProviderAzure         = "azure"
	azureRegion                = "westus2"
	awsRegion                  = "us-west-2"
	vaultClusterResourceName   = "hcp_vault_cluster.test"
	vaultClusterDataSourceName = "data.hcp_vault_cluster.test"
	adminTokenResourceName     = "hcp_vault_cluster_admin_token.test"
)

const vaultCluster = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "{{ .ClusterID }}"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "{{ .Tier }}"
}
`

// sets public_endpoint to true and add metrics and audit log
const updatedVaultClusterPublicAndMetricsAuditLog = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "{{ .ClusterID }}"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "{{ .Tier }}"
	public_endpoint    = {{ .PublicEndpoint }}
	metrics_config			   {
		splunk_hecendpoint = "https://http-input-splunkcloud.com"
		splunk_token =       "test"
	}
	audit_log_config 		    {
		datadog_api_key = "test_datadog"
		datadog_region  = "us1"
	}
	major_version_upgrade_config {
		upgrade_type = "MANUAL"
	}
}
`

// changes tier, remove any metrics or audit log config, optionally toggle public
// endpoint on or off
const updatedVaultClusterTierAndMVUConfig = `
resource "hcp_vault_cluster" "test" {
	cluster_id         = "{{ .ClusterID }}"
	hvn_id             = hcp_hvn.test.hvn_id
	tier               = "{{ .Tier }}"
	public_endpoint    = {{ .PublicEndpoint }}
	major_version_upgrade_config {
		upgrade_type = "SCHEDULED"
		maintenance_window_day = "WEDNESDAY"
		maintenance_window_time = "WINDOW_12AM_4AM"
	}
}
`

func setTestAccVaultClusterConfig(t *testing.T, tfCode string, in inputT, tier string) string {
	tfTemplate := fmt.Sprintf(`
resource "hcp_hvn" "test" {
	hvn_id            = "{{ .HvnID }}"
	cloud_provider    = "{{ .CloudProvider }}"
	region            = "{{ .Region }}"
}

%s

data "hcp_vault_cluster" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}

resource "hcp_vault_cluster_admin_token" "test" {
	cluster_id       = hcp_vault_cluster.test.cluster_id
}
`, tfCode)

	tmpl, err := template.New("tf_resources").Parse(tfTemplate)
	require.NoError(t, err)

	tfResources := &bytes.Buffer{}
	err = tmpl.Execute(tfResources, struct {
		ClusterID      string
		HvnID          string
		HvnCidr        string
		CloudProvider  string
		Region         string
		Tier           string
		PublicEndpoint string
	}{
		ClusterID:      in.VaultClusterName,
		HvnID:          in.HvnName,
		CloudProvider:  in.CloudProvider,
		Region:         in.Region,
		Tier:           tier,
		PublicEndpoint: in.PublicEndpoint,
	})
	require.NoError(t, err)
	return tfResources.String()
}
