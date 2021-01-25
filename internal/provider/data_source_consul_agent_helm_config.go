package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultConsulAgentHelmConfigTimeoutDuration is the default timeout
// for reading the agent Helm config.
var defaultConsulAgentHelmConfigTimeoutDuration = time.Minute * 5

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
				Description:      "The FQDN for the Kubernetes auth method endpoint.",
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
func dataSourceConsulAgentHelmConfigRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
