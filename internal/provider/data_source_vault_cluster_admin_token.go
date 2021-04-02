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

// adminTokenExpiry is the length of the time in seconds before a generated admin token expires.
var adminTokenExpiry = time.Second * 3600 * 6

func dataSourceVaultClusterAdminToken() *schema.Resource {
	return &schema.Resource{
		Description: "The Vault cluster admin token resource generates an admin-level token for the HCP Vault cluster.",
		ReadContext: dataSourceVaultClusterAdminTokenRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Vault cluster.",
				Type:             schema.TypeString,
				Required:         true,
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
				Sensitive:   true,
			},
		},
	}
}

// dataSourceVaultClusterAdminTokenRead always generates a new admin token for the Vault cluster.
func dataSourceVaultClusterAdminTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	log.Printf("[INFO] checking for existing admin token for Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	// TODO: this is not how I access existing state's created_at...
	createdAt := d.Get("created_at").(string)
	if createdAt != "" {
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
