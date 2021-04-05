package provider

import (
	"context"
	"fmt"
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
			"created_at": {
				Description: "The time that the admin token was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"token": {
				Description: "The admin token of this HCP Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// resourceVaultClusterAdminTokenCreate generates a new admin token for the Vault cluster if there is no existing token or the 
// current token in state is about to expire.
func resourceVaultClusterAdminTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	loc := &models.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	log.Printf("[INFO] reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
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

	createdAt, ok := d.GetOk("created_at").(string); ok {
		log.Printf("[INFO] existing admin token found for Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

		// If the token is less than five minutes from expiry, it's time to regenerate.
		expiry := adminTokenExpiry - (time.Second * 60 * 5)
		// TODO: for testing only
		// expiry := time.Second * 20

		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return diag.Errorf("error verifying HCP Vault cluster admin token (cluster_id %q) (project_id %q): %+v",
				clusterID,
				client.Config.ProjectID,
				err,
			)
		}

		// Return early if admin token is still within expiry window.
		if time.Now().Unix() < t.Add(expiry).Unix() {
			return nil
		}
	}

	loc.Region = &models.HashicorpCloudLocationRegion{
		Provider: cluster.Location.Region.Provider,
		Region:   cluster.Location.Region.Region,
	}

	tokenResp, err := clients.CreateVaultClusterAdminToken(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("error creating HCP Vault cluster admin token (cluster_id %q) (project_id %q): %+v",
			clusterID,
			client.Config.ProjectID,
			err,
		)
	}

	err = d.Set("token", tokenResp.Token)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("created_at", time.Now().Format(time.RFC3339))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("/project/%s/%s/%s/token",
		loc.ProjectID,
		VaultClusterResourceType,
		clusterID))

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
