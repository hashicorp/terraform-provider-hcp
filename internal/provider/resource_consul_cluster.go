package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/consul"
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
			// optional fields
			"public_endpoint": {
				Description: "Denotes that the cluster has an public endpoint for the Consul UI. Defaults to false.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"min_consul_version": {
				Description:      "The minimum Consul version of the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSemVer,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// Suppress diff is normalized versions match OR min_consul_version is removed from the resource
					// since min_consul_version is required in order to upgrade the cluster to a new Consul version.
					return consul.NormalizeVersion(old) == consul.NormalizeVersion(new) || new == ""
				},
			},
			"datacenter": {
				Description: "The Consul data center name of the cluster. If not specified, it is defaulted to the value of `id`.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"connect_enabled": {
				Description: "Denotes the Consul connect feature should be enabled for this cluster.  Default to true.",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			// computed outputs
			"state": {
				Description: "The state of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_automatic_upgrades": {
				Description: "Denotes that automatic Consul upgrades are enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"consul_snapshot_interval": {
				Description: "The Consul snapshot interval.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_snapshot_retention": {
				Description: "The retention policy for Consul snapshots.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_config_file": {
				Description: "The cluster config encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_ca_file": {
				Description: "The cluster CA file encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_connect": {
				Description: "Denotes that Consul connect is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"consul_version": {
				Description: "The Consul version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_public_endpoint_url": {
				Description: "The public URL for the Consul UI. This will be empty if `public_endpoint` is `true`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_private_endpoint_url": {
				Description: "The private URL for the Consul UI.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_root_token_accessor_id": {
				Description: "The accessor ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_root_token_secret_id": {
				Description: "The secret ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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
