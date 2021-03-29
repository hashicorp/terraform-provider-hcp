package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
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

func resourceVaultClusterAdminTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	// Fetch organizationID by project ID.
	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	_, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to create admin token; Vault cluster (%s) not found",
				clusterID,
			)
		}

		return diag.Errorf("unable to check for presence of an existing Vault cluster (%s): %v",
			clusterID,
			err,
		)
	}

	tokenResp, err := clients.CreateVaultClusterAdminToken(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("error creating HCP Vault cluster admin token (cluster_id %q) (project_id %q): %+v",
			clusterID,
			projectID,
			err,
		)
	}

	err = d.Set("token", tokenResp.Token)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: Need to verify if this is safe to be used as ID in state.
	d.SetId(tokenResp.Token)

	return nil
}

// resourceVaultClusterAdminTokenRead will act as a no-op as the admin token is not persisted in
// any way that it can be fetched and read
func resourceVaultClusterAdminTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	organizationID := client.Config.OrganizationID
	projectID := client.Config.ProjectID

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	_, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// No cluster exists, so this admin token should be removed from state.
			log.Printf("[WARN] no HCP Vault cluster found with (cluster_id %q) (project_id %q); removing admin token.",
				clusterID,
				projectID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing Vault cluster (cluster_id %q) (project_id %q): %v",
			clusterID,
			projectID,
			err,
		)
	}

	return nil
}

// resourceVaultClusterAdminTokenDelete will act as a no-op as the admin token is not persisted in
// any way that it can be deleted.
func resourceVaultClusterAdminTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVaultClusterAdminTokenRead(ctx, d, meta)
}
