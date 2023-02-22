// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultConsulAgentHelmConfigTimeoutDuration is the default timeout
// for reading the agent Helm config.
var defaultConsulAgentHelmConfigTimeoutDuration = time.Minute * 5

// ConsulConfig represents the Consul configuration that will be
// decoded from a base64 formatted string.
type ConsulConfig struct {
	Datacenter string   `json:"datacenter"`
	Encrypt    string   `json:"encrypt"`
	RetryJoin  []string `json:"retry_join"`
}

// helmConfigTemplate is the template used to generate a helm
// config for an AKS cluster based on given inputs.
//
// see generateHelmConfig for details on the inputs passed in
const helmConfigTemplate = `global:
  enabled: false
  name: consul
  datacenter: %s
  acls:
    manageSystemACLs: true
    bootstrapToken:
      secretName: %s-bootstrap-token
      secretKey: token
  gossipEncryption:
    secretName: %s-hcp
    secretKey: gossipEncryptionKey
  tls:
    enabled: true
    enableAutoEncrypt: true
    caCert:
      secretName: %s-hcp
      secretKey: caCert
externalServers:
  enabled: true
  hosts: %s
  httpsPort: 443
  useSystemRoots: true
  k8sAuthMethodHost: https://%s:443
client:
  enabled: true
  exposeGossipPorts: %t
  join: %s
connectInject:
  enabled: true`

func dataSourceConsulAgentHelmConfig() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.",
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultConsulAgentHelmConfigTimeoutDuration,
		},
		ReadContext: dataSourceConsulAgentHelmConfigRead,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"kubernetes_endpoint": {
				Description:      "The FQDN for the Kubernetes API.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional
			"expose_gossip_ports": {
				Description: "Denotes that the gossip ports should be exposed.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// Computed outputs
			"config": {
				Description: "The agent Helm config.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceConsulAgentHelmConfigRead is the func to implement reading of the
// Consul agent Helm config for an HCP cluster.
func dataSourceConsulAgentHelmConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	clusterID := d.Get("cluster_id").(string)

	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	v, ok := d.GetOk("project_id")
	if ok {
		projectID = v.(string)
	}

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to read Consul agent Helm config; Consul cluster (%s) not found",
				clusterID,
			)

		}

		return diag.Errorf("unable to check for presence of an existing Consul cluster (%s): %v",
			clusterID,
			err,
		)
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
		return diag.FromErr(fmt.Errorf("unable to base64 decode Consul config (%v): %v", configStr, err))
	}

	// unmarshal from JSON
	var consulConfig ConsulConfig
	err = json.Unmarshal(consulConfigJSON, &consulConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to json unmarshal consul config %v", err))
	}

	// generate helm config and set on data source
	if err := d.Set("config", generateHelmConfig(
		cluster.ID,
		cluster.Config.ConsulConfig.Datacenter,
		d.Get("kubernetes_endpoint").(string),
		consulConfig.RetryJoin,
		d.Get("expose_gossip_ports").(bool))); err != nil {
		return diag.FromErr(err)
	}

	// build ID and set it
	link := newLink(loc, ConsulClusterHelmConfigDataSourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	return nil
}

// generateHelmConfig will generate a helm config based on the passed in
// name, data center, retry join, fqdn and expose gossip ports flag.
func generateHelmConfig(name, datacenter, fqdn string, retryJoin []string, exposeGossipPorts bool) string {
	// lowercase the name
	lower := strings.ToLower(name)

	// print retryJoin a double-quoted string safely escaped with Go syntax
	rj := fmt.Sprintf("%q", retryJoin)

	// replace any escaped double-quotes with single quotes
	rj = strings.ReplaceAll(rj, "\"", "'")

	// trim off any leading `https://` protocol if present.
	// this protocol will be prepended as expected when
	// the helm config string is generated.
	//
	// this string is trimmed here to handle both the cases
	// of when the provided fqdn string has the leading
	// protocol and does not.
	//
	// trimming the leading protocol here will guarantee a
	// valid URL is generated when prepended in the
	// helmConfigTemplate.
	fqdn = strings.TrimPrefix(fqdn, "https://")

	return fmt.Sprintf(helmConfigTemplate,
		datacenter,
		lower, lower, lower,
		rj,
		fqdn,
		exposeGossipPorts,
		rj,
	)
}
