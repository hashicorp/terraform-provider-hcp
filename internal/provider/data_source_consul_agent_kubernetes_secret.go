// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultAgentConfigKubernetesSecretTimeoutDuration is the default timeout
// for reading the agent config Kubernetes secret.
var defaultAgentConfigKubernetesSecretTimeoutDuration = time.Minute * 5

// agentConfigKubernetesSecretTemplate is the template used to generate a
// Kubernetes formatted secret for the Consul agent config.
const agentConfigKubernetesSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: %s-hcp
type: Opaque
data:
  gossipEncryptionKey: %s
  caCert: %s`

func dataSourceConsulAgentKubernetesSecret() *schema.Resource {
	return &schema.Resource{
		Description: "The agent config Kubernetes secret data source provides Consul agents running in Kubernetes the configuration needed to connect to the Consul cluster.",
		ReadContext: dataSourceConsulAgentKubernetesSecretRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultAgentConfigKubernetesSecretTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"secret": {
				Description: "The Consul agent configuration in the format of a Kubernetes secret (YAML).",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceAgentConfigKubernetesSecretRead retrieves the Consul config and formats a Kubernetes secret for Consul agents running
// in Kubernetes to leverage.
func dataSourceConsulAgentKubernetesSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	projectID := client.Config.ProjectID
	organizationID := client.Config.OrganizationID

	clusterID := d.Get("cluster_id").(string)

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s) client config files: %v", clusterID, err)
	}

	// pull off the config string
	configStr := clientConfigFiles.ConsulConfigFile.String()

	// decode it
	consulConfigJSON, err := base64.StdEncoding.DecodeString(configStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to base64 decode consul config (%v): %w", configStr, err))
	}

	// unmarshal from JSON
	var consulConfig ConsulConfig
	err = json.Unmarshal(consulConfigJSON, &consulConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to json unmarshal consul config %v", err))
	}

	encodedGossipKey := base64.StdEncoding.EncodeToString([]byte(consulConfig.Encrypt))
	encodedCAFile := clientConfigFiles.CaFile

	err = d.Set("secret", fmt.Sprintf(agentConfigKubernetesSecretTemplate, clusterID, encodedGossipKey, encodedCAFile))
	if err != nil {
		return diag.FromErr(err)
	}

	// build ID and set it
	link := newLink(loc, ConsulClusterAgentKubernetesSecretDataSourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	return nil
}
