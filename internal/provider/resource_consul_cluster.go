package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultClusterTimeoutDuration is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultClusterTimeoutDuration = time.Minute * 5

// createUpdateTimeoutDuration is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateTimeoutDuration = time.Minute * 30

// deleteTimeoutDuration is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteTimeoutDuration = time.Minute * 25

// consulCusterResourceCloudProviders is the list of cloud providers
// where a HCP Consul cluster can be provisioned.
var consulCusterResourceCloudProviders = []string{
	"aws",
}

// consulClusterResourceTierLevels is the list of different tier
// levels that an HCP Consul cluster can be as.
var consulClusterResourceTierLevels = []string{
	"Development",
	"Production",
}

// resourceConsulCluster represents an HCP Consul cluster.
func resourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The Consul cluster resource allow you to manage an HCP Consul cluster.",
		CreateContext: resourceConsulClusterCreate,
		ReadContext:   resourceConsulClusterRead,
		UpdateContext: resourceConsulClusterUpdate,
		DeleteContext: resourceConsulClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterTimeoutDuration,
			Create:  &createUpdateTimeoutDuration,
			Update:  &createUpdateTimeoutDuration,
			Delete:  &deleteTimeoutDuration,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceConsulClusterImport,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"hvn_id": {
				Description:      "The ID of the HVN this HCP Consul cluster is associated to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"cluster_tier": {
				Description:      "The cluster tier of this HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSlice(consulClusterResourceTierLevels, true),
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			"cloud_provider": {
				Description:      "The provider where the HCP Consul cluster is located. Only 'aws' is available at this time.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSlice(consulCusterResourceCloudProviders, true),
			},
			"region": {
				Description:      "The region where the HCP Consul cluster is located.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
		},
	}
}

func resourceConsulClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
