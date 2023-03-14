// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"time"

	boundarymodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-boundary-service/stable/2021-12-21/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultBoundaryClusterTimeout = time.Minute * 5

// createUpdateBoundaryClusterTimeout is the amount of time that can elapse
// before a cluster create operation should timeout.
var createBoundaryClusterTimeout = time.Minute * 25

// deleteBoundaryClusterTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteBoundaryClusterTimeout = time.Minute * 25

func resourceBoundaryCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource allows you to manage an HCP Boundary cluster",
		CreateContext: resourceBoundaryClusterCreate,
		ReadContext:   resourceBoundaryClusterRead,
		DeleteContext: resourceBoundaryClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBoundaryClusterImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  &createBoundaryClusterTimeout,
			Delete:  &deleteBoundaryClusterTimeout,
			Default: &defaultBoundaryClusterTimeout,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"cluster_id": {
				Description:      "The ID of the Boundary cluster",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"username": {
				Description:      "The username of the initial admin user. This must be at least 3 characters in length, alphanumeric, hyphen, or period.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateBoundaryUsername,
			},
			"password": {
				Description:      "The password of the initial admin user. This must be at least 8 characters in length. Note that this may show up in logs, and it will be stored in the state file.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateBoundaryPassword,
				Sensitive:        true,
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

func resourceBoundaryClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	// gather the required bits to create a boundary cluster create request
	clusterID := d.Get("cluster_id").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
		// This is currently hardcoded, depending on decisions from PM
		// around regionality this may have to turn into an input
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: "aws",
			Region:   "us-east-1",
		},
	}

	// check for an existing boundary cluster
	_, err := clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Boundary cluster (%s): %v", clusterID, err)
		}
		// A 404 indicates a Boundary cluster was not found.
		log.Printf("[INFO] Boundary cluster (%s) not found, proceeding with create", clusterID)
	} else {
		return diag.Errorf("a Boundary cluster with cluster_id=%q in project_id=%q already exists.", clusterID, loc.ProjectID)
	}

	// assemble the BoundaryClusterCreateRequest
	req := &boundarymodels.HashicorpCloudBoundary20211221CreateRequest{
		ClusterID: clusterID,
		Username:  username,
		Password:  password,
		Location:  loc,
	}

	// execute the Boundary cluster creation
	log.Printf("[INFO] Creating Boundary cluster (%s)", clusterID)
	createResp, err := clients.CreateBoundaryCluster(ctx, client, loc, req)
	if err != nil {
		return diag.Errorf("unable to create Boundary cluster (%s): %v", clusterID, err)
	}
	link := newLink(loc, BoundaryClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for the Boundary cluster to be created.
	if err := clients.WaitForOperation(ctx, client, "create Boundary cluster", loc, createResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create Boundary cluster (%s): %v", createResp.ClusterID, err)
	}
	log.Printf("[INFO] Created Boundary cluster (%s)", createResp.ClusterID)

	// Get the created Boundary cluster.
	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, createResp.ClusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Boundary cluster (%s): %v", createResp.ClusterID, err)
	}
	// set Boundary cluster resource data
	err = setBoundaryClusterResourceData(d, cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBoundaryClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), BoundaryClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Boundary cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetBoundaryClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Boundary cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to fetch Boundary cluster (%s): %v", clusterID, err)
	}

	// The Boundary cluster was already deleted, remove from state.
	if *cluster.State == boundarymodels.HashicorpCloudBoundary20211221ClusterStateSTATEDELETED {
		log.Printf("[WARN] Boundary cluster (%s) failed to provision, removing from state", clusterID)
		d.SetId("")
		return nil
	}

	// Cluster found, update resource data.
	if err := setBoundaryClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBoundaryClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), BoundaryClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Deleting Boundary cluster (%s)", clusterID)

	deleteResp, err := clients.DeleteBoundaryCluster(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Boundary cluster (%s) not found, so no action was taken", clusterID)
			return nil
		}

		return diag.Errorf("unable to delete Boundary cluster (%s): %v", clusterID, err)
	}

	// Wait for the delete cluster operation.
	if err := clients.WaitForOperation(ctx, client, "delete Boundary cluster", loc, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Boundary cluster (%s): %v", clusterID, err)
	}

	return nil
}

func resourceBoundaryClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	clusterID := d.Id()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	link := newLink(loc, BoundaryClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}

func setBoundaryClusterResourceData(d *schema.ResourceData, cluster *boundarymodels.HashicorpCloudBoundary20211221Cluster) error {
	if err := d.Set("cluster_id", cluster.ClusterID); err != nil {
		return err
	}
	createdAtStr := cluster.CreatedAt.String()
	if err := d.Set("created_at", createdAtStr); err != nil {
		return err
	}
	if err := d.Set("cluster_url", cluster.ClusterURL); err != nil {
		return err
	}
	if err := d.Set("state", cluster.State); err != nil {
		return err
	}
	return nil
}
