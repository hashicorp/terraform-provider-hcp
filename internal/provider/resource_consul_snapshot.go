// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"strconv"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

const (
	// defaultRestoredAt is the default string returned when a snapshot has not been restored
	defaultRestoredAt = "0001-01-01T00:00:00.000Z"
)

// defaultSnapshotTimeoutDuration is the amount of time that can elapse
// before a snapshot read should timeout.
var defaultSnapshotTimeoutDuration = time.Minute * 5

// snapshotCreateUpdateDeleteTimeoutDuration is the amount of time that can elapse
// before a snapshot operation should timeout.
var snapshotCreateUpdateDeleteTimeoutDuration = time.Minute * 15

func resourceConsulSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul snapshot resource allows users to manage Consul snapshots of an HCP Consul cluster. " +
			"Snapshots currently have a retention policy of 30 days.",
		CreateContext: resourceConsulSnapshotCreate,
		ReadContext:   resourceConsulSnapshotRead,
		UpdateContext: resourceConsulSnapshotUpdate,
		DeleteContext: resourceConsulSnapshotDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  &snapshotCreateUpdateDeleteTimeoutDuration,
			Update:  &snapshotCreateUpdateDeleteTimeoutDuration,
			Delete:  &snapshotCreateUpdateDeleteTimeoutDuration,
			Default: &defaultSnapshotTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"snapshot_name": {
				Description:      "The name of the snapshot.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// computed outputs
			"project_id": {
				Description: "The ID of the project the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"snapshot_id": {
				Description: "The ID of the Consul snapshot",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization_id": {
				Description: "The ID of the HCP organization where the project the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The size of the snapshot in bytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"consul_version": {
				Description: "The version of Consul at the time of snapshot creation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"restored_at": {
				Description: "Timestamp of when the snapshot was restored. If the snapshot has not been restored, this field will be blank.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of an HCP Consul snapshot.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConsulSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	// Check for an existing Consul cluster
	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Consul cluster (%s): %v", clusterID, err)
		}

		// a 404 indicates a Consul cluster was not found
		return diag.Errorf("unable to create snapshot; no HCP Cluster found for Consul cluster (%s)", clusterID)
	}

	name := d.Get("snapshot_name").(string)

	log.Printf("[INFO] Creating Consul snapshot (%s)", name)

	// make the call to kick off the workflow
	createResp, err := clients.CreateSnapshot(ctx, client, newLink(cluster.Location, ConsulClusterResourceType, cluster.ID), name)
	if err != nil {
		return diag.Errorf("unable to create Consul snapshot (%s): %v", clusterID, err)
	}

	log.Printf("[INFO] Created Consul snapshot name:%q; id:%q", name, createResp.SnapshotID)

	link := newLink(loc, ConsulSnapshotResourceType, createResp.SnapshotID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// set the consul_version based on the cluster's Consul version
	err = d.Set("consul_version", cluster.ConsulVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	// wait for the Consul snapshot to be created
	if err := clients.WaitForOperation(ctx, client, ConsulSnapshotResourceType+".create", cluster.Location, createResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", cluster.ID, err)
	}

	// return resourceSnapshotRead
	return resourceConsulSnapshotRead(ctx, d, meta)
}

func resourceConsulSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	snapshotLink, err := buildLinkFromURL(d.Id(), ConsulSnapshotResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID := snapshotLink.ID
	loc := snapshotLink.Location

	snapshotResp, err := clients.GetSnapshotByID(ctx, client, loc, snapshotID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul snapshot (%s) not found, removing from state", snapshotID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Consul snapshot (%s): %v", snapshotID, err)
	}

	if err := setConsulSnapshotResourceData(d, snapshotResp.Snapshot); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConsulSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	snapshotLink, err := buildLinkFromURL(d.Id(), ConsulSnapshotResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID := snapshotLink.ID
	loc := snapshotLink.Location

	snapshot, err := clients.GetSnapshotByID(ctx, client, loc, snapshotID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul snapshot (%s) not found, removing from state", snapshotID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Consul snapshot (%s): %v", snapshotID, err)
	}

	name := d.Get("snapshot_name").(string)
	updateResp, err := clients.RenameSnapshotByID(ctx, client, loc, snapshot.Snapshot.ID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setConsulSnapshotResourceData(d, updateResp.Snapshot); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConsulSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), ConsulSnapshotResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID := link.ID
	loc := link.Location

	log.Printf("[INFO] Deleting Consul snapshot (%s)", snapshotID)

	deleteResp, err := clients.DeleteSnapshotByID(ctx, client, loc, snapshotID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul snapshot (%s) not found, so no action was taken", snapshotID)
			return nil
		}

		return diag.Errorf("unable to delete Consul snapshot (%s): %v", snapshotID, err)
	}

	// Wait for the delete snapshot operation
	if err := clients.WaitForOperation(ctx, client, ConsulSnapshotResourceType+".delete", loc, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Consul snapshot (%s): %v", snapshotID, err)
	}

	log.Printf("[INFO] Consul snapshot (%s) deleted, removing from state", snapshotID)

	return nil
}

func setConsulSnapshotResourceData(d *schema.ResourceData, snapshot *consulmodels.HashicorpCloudConsul20210204Snapshot) error {

	if err := d.Set("organization_id", snapshot.Location.OrganizationID); err != nil {
		return err
	}

	if err := d.Set("project_id", snapshot.Location.ProjectID); err != nil {
		return err
	}

	if err := d.Set("snapshot_id", snapshot.ID); err != nil {
		return err
	}

	if err := d.Set("state", snapshot.State); err != nil {
		return err
	}

	if snapshot.Meta != nil {
		size, err := strconv.Atoi(snapshot.Meta.Size)
		if err != nil {
			return err
		}
		if err := d.Set("size", size); err != nil {
			return err
		}

		if snapshot.Meta.RestoredAt.String() != defaultRestoredAt {
			if err := d.Set("restored_at", snapshot.Meta.RestoredAt.String()); err != nil {
				return err
			}
		}
	}
	return nil
}
