// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// This string is used to represent an unassigned (or "null") channel
// assignment for iteration identifiers of the String type
const unassignString string = "none"

func resourcePackerChannelAssignment() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Channel Assignment resource allows you to manage the iteration assigned to a bucket channel in an active HCP Packer Registry.",
		CreateContext: resourcePackerChannelAssignmentCreate,
		DeleteContext: resourcePackerChannelAssignmentDelete,
		ReadContext:   resourcePackerChannelAssignmentRead,
		UpdateContext: resourcePackerChannelAssignmentUpdate,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultPackerTimeout,
			Default: &defaultPackerTimeout,
			Update:  &defaultPackerTimeout,
			Delete:  &defaultPackerTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourcePackerChannelAssignmentImport,
		},
		CustomizeDiff: resourcePackerChannelAssignmentCustomizeDiff,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"channel_name": {
				Description:      "The name of the HCP Packer channel being managed.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"bucket_name": {
				Description:      "The slug of the HCP Packer Registry bucket where the channel is located.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the channel is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"iteration_fingerprint": {
				Description:  "The fingerprint of the iteration assigned to the channel.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "iteration_fingerprint", "iteration_version"},
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"iteration_id": {
				Description:  "The ID of the iteration assigned to the channel.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "iteration_fingerprint", "iteration_version"},
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"iteration_version": {
				Description:  "The incremental version of the iteration assigned to the channel.",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "iteration_fingerprint", "iteration_version"},
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
			},
			// Computed Values
			"organization_id": {
				Description: "The ID of the HCP organization where this channel is located. Always the same as the associated channel.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePackerChannelAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	channel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel == nil {
		d.SetId("")
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found.", channelName, bucketName, loc.ProjectID),
		}}
	}

	if err := setPackerChannelAssignmentIterationData(d, channel.Iteration); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackerChannelAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	channel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	} else if channel == nil {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found.", channelName, bucketName, loc.ProjectID),
		}}
	} else if channel.Managed {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) is managed by HCP Packer and cannot have an iteration assigned by Terraform.", channelName, bucketName, loc.ProjectID),
		}}
	} else if iteration := channel.Iteration; iteration != nil && (iteration.IncrementalVersion > 0 || iteration.ID != "" || iteration.Fingerprint != "") {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) already has an assigned iteration.", channelName, bucketName, loc.ProjectID),
			Detail:   "To adopt this resource into Terraform, use `terraform import`, or remove the channel's assigned iteration using the HCP Packer GUI/API.",
		}}
	}

	iterID := d.Get("iteration_id").(string)
	if iterID == unassignString {
		iterID = ""
	}

	iterFingerprint := d.Get("iteration_fingerprint").(string)
	if iterFingerprint == unassignString {
		iterFingerprint = ""
	}

	updatedChannel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName,
		&packermodels.HashicorpCloudPackerIteration{
			IncrementalVersion: int32(d.Get("iteration_version").(int)),
			ID:                 iterID,
			Fingerprint:        iterFingerprint,
		}, nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(updatedChannel.ID)

	if err := setPackerChannelAssignmentIterationData(d, updatedChannel.Iteration); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackerChannelAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	iteration := &packermodels.HashicorpCloudPackerIteration{}
	assignmentHasChanges := d.HasChanges("iteration_version", "iteration_id", "iteration_fingerprint")
	if !assignmentHasChanges || d.HasChange("iteration_version") {
		iteration.IncrementalVersion = int32(d.Get("iteration_version").(int))
	}
	if iterID := d.Get("iteration_id").(string); (!assignmentHasChanges || d.HasChange("iteration_id")) && iterID != unassignString {
		iteration.ID = iterID
	}
	if iterFingerpint := d.Get("iteration_fingerprint").(string); (!assignmentHasChanges || d.HasChange("iteration_fingerprint")) && iterFingerpint != unassignString {
		iteration.Fingerprint = iterFingerpint
	}

	updatedChannel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName, iteration, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setPackerChannelAssignmentIterationData(d, updatedChannel.Iteration); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackerChannelAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	_, err = clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName, nil, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackerChannelAssignmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_packer_channel_assignment.test {project_id}:{bucket_name}:{channel_name}
	// use default project ID from provider:
	//   terraform import hcp_packer_channel_assignment.test {bucket_name}:{channel_name}

	client := meta.(*clients.Client)
	bucketName := ""
	channelName := ""
	projectID := ""
	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	idParts := strings.SplitN(d.Id(), ":", 3)
	if len(idParts) == 3 { // {project_id}:{bucket_name}:{channel_name}
		if idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {project_id}:{bucket_name}:{channel_name}", d.Id())
		}
		projectID = idParts[0]
		bucketName = idParts[1]
		channelName = idParts[2]
	} else if len(idParts) == 2 { // {bucket_name}:{channel_name}
		if idParts[0] == "" || idParts[1] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {bucket_name}:{channel_name}", d.Id())
		}
		projectID, err = GetProjectID(projectID, client.Config.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve project ID: %v", err)
		}
		bucketName = idParts[0]
		channelName = idParts[1]
	} else {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected {bucket_name}:{channel_name} or {project_id}:{bucket_name}:{channel_name}", d.Id())
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	if err := setLocationResourceData(d, loc); err != nil {
		return nil, err
	}

	if err := d.Set("bucket_name", bucketName); err != nil {
		return nil, err
	}
	if err := d.Set("channel_name", channelName); err != nil {
		return nil, err
	}

	channel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return nil, err
	} else if channel == nil {
		return nil, fmt.Errorf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found", channelName, bucketName, loc.ProjectID)
	} else if channel.Managed {
		return nil, fmt.Errorf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) is managed by HCP Packer and cannot have an iteration assigned by Terraform", channelName, bucketName, loc.ProjectID)
	}

	d.SetId(channel.ID)

	if err := setPackerChannelAssignmentIterationData(d, channel.Iteration); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePackerChannelAssignmentCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	client := meta.(*clients.Client)
	var err error
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return fmt.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	bucketName := d.Get("bucket_name").(string)

	if (d.HasChange("iteration_id") && !d.NewValueKnown("iteration_id")) ||
		(d.HasChange("iteration_fingerprint") && !d.NewValueKnown("iteration_fingerprint")) ||
		(d.HasChanges("iteration_version") && !d.NewValueKnown("iteration_id")) {
		if err := d.SetNewComputed("iteration_id"); err != nil {
			return err
		}
		if err := d.SetNewComputed("iteration_fingerprint"); err != nil {
			return err
		}
		if err := d.SetNewComputed("iteration_version"); err != nil {
			return err
		}
	} else {
		var iteration *packermodels.HashicorpCloudPackerIteration
		var itErr error

		if rawID, ok := d.GetOk("iteration_id"); ok && d.HasChange("iteration_id") && d.NewValueKnown("iteration_id") {
			if id := rawID.(string); id != unassignString {
				iteration, itErr = clients.GetIterationFromID(ctx, client, loc, bucketName, id)
			} else {
				iteration = &packermodels.HashicorpCloudPackerIteration{}
			}
		} else if rawFingerprint, ok := d.GetOk("iteration_fingerprint"); ok && d.HasChange("iteration_fingerprint") && d.NewValueKnown("iteration_fingerprint") {
			if fingerprint := rawFingerprint.(string); fingerprint != unassignString {
				iteration, itErr = clients.GetIterationFromFingerprint(ctx, client, loc, bucketName, fingerprint)
			} else {
				iteration = &packermodels.HashicorpCloudPackerIteration{}
			}
		} else if rawVersion, ok := d.GetOk("iteration_version"); ok && d.HasChange("iteration_version") && d.NewValueKnown("iteration_version") {
			if version := int32(rawVersion.(int)); version != 0 {
				iteration, itErr = clients.GetIterationFromVersion(ctx, client, loc, bucketName, version)
			} else {
				iteration = &packermodels.HashicorpCloudPackerIteration{}
			}
		}

		if itErr != nil {
			return itErr
		} else if iteration != nil {
			if err := d.SetNew("iteration_version", iteration.IncrementalVersion); err != nil {
				return err
			}

			id := iteration.ID
			if id == "" {
				id = unassignString
			}
			if err := d.SetNew("iteration_id", id); err != nil {
				return err
			}

			fingerprint := iteration.Fingerprint
			if fingerprint == "" {
				fingerprint = unassignString
			}
			if err := d.SetNew("iteration_fingerprint", fingerprint); err != nil {
				return err
			}
		}
	}

	return nil
}

func setPackerChannelAssignmentIterationData(d *schema.ResourceData, i *packermodels.HashicorpCloudPackerIteration) error {
	var iteration packermodels.HashicorpCloudPackerIteration

	if i == nil {
		iteration = packermodels.HashicorpCloudPackerIteration{
			IncrementalVersion: 0,
			ID:                 "",
			Fingerprint:        "",
		}
	} else {
		iteration = *i
	}

	if err := d.Set("iteration_version", iteration.IncrementalVersion); err != nil {
		return err
	}

	id := iteration.ID
	if id == "" {
		id = unassignString
	}
	if err := d.Set("iteration_id", id); err != nil {
		return err
	}

	fingerprint := iteration.Fingerprint
	if fingerprint == "" {
		fingerprint = unassignString
	}
	if err := d.Set("iteration_fingerprint", fingerprint); err != nil {
		return err
	}

	return nil
}
