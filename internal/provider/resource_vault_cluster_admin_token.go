package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// defaultVaultAdminTokenTimeout is the amount of time that can elapse
// before an admin token create operation should timeout.
var defaultVaultAdminTokenTimeout = time.Minute * 5

func resourceVaultClusterAdminToken() *schema.Resource {
	return &schema.Resource{
		Description:   "The Vault cluster admin token resource provides a token with administrator privileges on an HCP Vault cluster.",
		CreateContext: resourceVaultClusterAdminTokenCreate,
		ReadContext:   resourceVaultClusterAdminTokenRead,
		DeleteContext: resourceVaultClusterAdminTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: &defaultVaultAdminTokenTimeout,
			Read:   &defaultVaultAdminTokenTimeout,
			Delete: &defaultVaultAdminTokenTimeout,
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
			// computed outputs
			"token": {
				Description: "The admin token of this HCP Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVaultClusterAdminTokenCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceVaultClusterAdminTokenRead will act as a no-op as the admin token is not persisted in
// any way that it can be fetched and read
func resourceVaultClusterAdminTokenRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceVaultClusterAdminTokenDelete will act as a no-op as the admin token is not persisted in
// any way that it can be deleted.
func resourceVaultClusterAdminTokenDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
