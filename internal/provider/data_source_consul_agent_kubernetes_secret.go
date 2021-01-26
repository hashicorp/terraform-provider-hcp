package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultAgentConfigKubernetesSecretTimeoutDuration is the default timeout
// for reading the agent config Kubernetes secret.
var defaultAgentConfigKubernetesSecretTimeoutDuration = time.Minute * 5

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
func dataSourceConsulAgentKubernetesSecretRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
