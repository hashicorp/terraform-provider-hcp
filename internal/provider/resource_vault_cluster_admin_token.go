// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
// before an admin token operation should timeout.
var defaultVaultAdminTokenTimeout = time.Minute * 5

// adminTokenExpiry is the length of the time in seconds before a generated admin token expires.
var adminTokenExpiry = time.Second * 3600 * 6

func resourceVaultClusterAdminToken() *schema.Resource {
	return &schema.Resource{
		Description:   "The Vault cluster admin token resource generates an admin-level token for the HCP Vault cluster.",
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
				Sensitive:   true,
			},
		},
	}
}

// resourceVaultClusterAdminTokenCreate generates a new admin token for the Vault cluster.
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

// resourceVaultClusterAdminTokenRead cannot read the admin token from the API as it is not persisted in
// any way that it can be fetched. Instead this operation first verifies the existence of the associated Vault cluster
// and then refreshes the token if it is close to expiring or expired.
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

	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
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

	loc.Region = &models.HashicorpCloudLocationRegion{
		Provider: cluster.Location.Region.Provider,
		Region:   cluster.Location.Region.Region,
	}

	// If the token already exists, this block verifies if it is close to expiration and should be refreshed.
	createdAt := d.Get("created_at").(string)
	if createdAt != "" {
		log.Printf("[INFO] existing admin token found for Vault cluster (%s) [project_id=%s, organization_id=%s]",
			clusterID,
			loc.ProjectID,
			loc.OrganizationID,
		)

		// The refresh window starts five minutes before the 6h expiry.
		expiry := adminTokenExpiry - (time.Second * 60 * 5)

		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return diag.Errorf("error verifying HCP Vault cluster admin token (cluster_id %q) (project_id %q): %+v",
				clusterID,
				client.Config.ProjectID,
				err,
			)
		}

		// If the token is less than five minutes from the 6h expiry, it's time to regenerate.
		if time.Now().Unix() > t.Add(expiry).Unix() {
			log.Printf("[INFO] refreshing admin token for Vault cluster (%s) [project_id=%s, organization_id=%s]",
				clusterID,
				loc.ProjectID,
				loc.OrganizationID,
			)

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
		}
	}

	return nil
}

// resourceVaultClusterAdminTokenDelete will remove the token from state but there is currently no way to invalidate an existing token.
func resourceVaultClusterAdminTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
