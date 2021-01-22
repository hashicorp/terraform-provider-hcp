package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceConsulSnapshotCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulSnapshotRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulSnapshotUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulSnapshotDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
