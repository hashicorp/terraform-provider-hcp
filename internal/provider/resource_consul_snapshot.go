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
