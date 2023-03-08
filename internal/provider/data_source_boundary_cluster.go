// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceBoundaryCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The Boundary cluster data source provides information about an existing HCP Boundary cluster.",
		ReadContext: dataSourceBoundaryClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultBoundaryClusterTimeout,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"cluster_id": {
				Description:      "The ID of the Boundary cluster",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"project_id": {
				Description:  "The ID of the HCP project where the Boundary cluster is located.",
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IsUUID,
			},
			// computed outputs
			"created_at": {
				Description: "The time that the Boundary cluster was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_url": {
				Description: "A unique URL identifying the Boundary cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the Boundary cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceBoundaryClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)
	clusterID := d.Get("cluster_id").(string)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] Reading Boundary cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	// get the boundary cluster by provided ID
	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	// build the id for this boundary cluster
	link := newLink(loc, BoundaryClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// cluster found, update resource data
	if err := setBoundaryClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
