package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultVaultClusterTimeout = time.Minute * 5

// createUpdateTimeout is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateVaultClusterTimeout = time.Minute * 35

// deleteTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteVaultClusterTimeout = time.Minute * 25

func resourceVaultCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The Vault cluster resource allows you to manage an HCP Vault cluster.",
		CreateContext: resourceVaultClusterCreate,
		ReadContext:   resourceVaultClusterRead,
		UpdateContext: resourceVaultClusterUpdate,
		DeleteContext: resourceVaultClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  &createUpdateVaultClusterTimeout,
			Update:  &createUpdateVaultClusterTimeout,
			Delete:  &deleteVaultClusterTimeout,
			Default: &defaultVaultClusterTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Vault cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"hvn_id": {
				Description:      "The ID of the HVN this HCP Vault cluster is associated to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"tier": {
				Description: "The tier that the HCP Vault cluster will be provisioned as.  Only 'development' and 'standard' are available at this time.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				// TODO: generalize consul enum validator
				// ValidateDiagFunc: validateStringInSlice(vaultClusterResourceTiers, true),
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			// optional fields
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint for the Vault UI. Defaults to false.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"namespace": {
				Description: "The name of the customer namespace.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"max_lease_ttl": {
				Description: "The max lease time-to-live (TTL) for this Vault cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"initial_vault_version": {
				Description:      "The initial Vault version to use when creating the cluster. Once the cluster is created, this value is no longer used. If not specified, it is defaulted to the version that is currently recommended by HCP.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSemVer,
				// DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
				// TODO generalize consul.NormalizeVersion()
				// },
			},
			// computed outputs
			"organization_id": {
				Description: "The ID of the organization this HCP Vault cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Vault cluster is located in.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"cloud_provider": {
				Description: "The provider where the HCP Vault cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "The region where the HCP Vault cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault_version": {
				Description: "The Vault version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault_public_endpoint_url": {
				Description: "The public URL for the Vault UI. This will be empty if `public_endpoint` is `false`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault_private_endpoint_url": {
				Description: "The private URL for the Vault UI.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the Vault cluster was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVaultClusterCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVaultClusterRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVaultClusterUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVaultClusterDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
