package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/helper"
)

// defaultSnapshotTimeoutDuration is the amount of time that can elapse
// before a snapshot read should timeout.
var defaultSnapshotTimeoutDuration = time.Minute * 5

// snapshotCreateUpdateDeleteTimeoutDuration is the amount of time that can elapse
// before a snapshot operation should timeout.
var snapshotCreateUpdateDeleteTimeoutDuration = time.Minute * 15

func resourceConsulSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul snapshot resource allows users to managed Consul snapshots of an HCP Consul cluster. " +
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
			// optional fields
			"project_id": {
				Description: "The ID of the project the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			// computed outputs
			"snapshot_id": {
				Description: "The ID of the Consul snapshot",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the snapshot.",
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
		},
	}
}

func resourceConsulSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	loc, err := helper.BuildResourceLocation(ctx, d, client, "Consul cluster")
	if err != nil {
		return diag.FromErr(err)
	}

	// Check for an existing Consul cluster
	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Consul Cluster (%s): %v", clusterID, err)
		}

		// a 404 indicates a Consul cluster was not found
		return diag.Errorf("unable to create snapshot; no HCS Cluster found for Consul cluster (%s)", clusterID)
	}

	name := d.Get("snapshot_name").(string)

	log.Printf("[INFO] Creating Consul snapshot (%s)", name)

	// make the call to kick off the workflow
	createResp, err := clients.CreateSnapshot(ctx, client, newLink(cluster.Location, "hashicorp.consul.cluster", cluster.ID), name)
	if err != nil {
		return diag.Errorf("unable to create Consul snapshot (%s): %v", clusterID, err)
	}

	log.Printf("[INFO] Created Consul snapshot name:%q; id:%q", name, createResp.SnapshotID)

	link := newLink(loc, "hashicorp.consul.snapshot", createResp.SnapshotID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// wait for the Consul snapshot to be created
	if err := clients.WaitForOperation(ctx, client, "create Consul cluster", cluster.Location, createResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", cluster.ID, err)
	}

	// return resourceSnapshotRead
	return resourceConsulSnapshotRead(ctx, d, meta)
}

func resourceConsulSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
