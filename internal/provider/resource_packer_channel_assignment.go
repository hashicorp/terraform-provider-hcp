// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

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
				Description:  "The ID of the HCP project where this channel is located. Must be the same as the associated channel.",
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
			},
			"iteration_id": {
				Description:  "The ID of the iteration assigned to the channel.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "iteration_fingerprint", "iteration_version"},
			},
			"iteration_version": {
				Description:  "The incremental version of the iteration assigned to the channel.",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "iteration_fingerprint", "iteration_version"},
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
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	if err := setLocationData(d, loc); err != nil {
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
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	if err := setLocationData(d, loc); err != nil {
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
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) is managed by HCP packer and cannot have an iteration assigned by Terraform.", channelName, bucketName, loc.ProjectID),
		}}
	} else if iteration := channel.Iteration; iteration != nil && (iteration.IncrementalVersion > 0 || iteration.ID != "" || iteration.Fingerprint != "") {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) already has an assigned iteration.", channelName, bucketName, loc.ProjectID),
			Detail:   "To adopt this resource into Terraform, use `terraform import`, or remove the channel's assigned iteration using the HCP Packer GUI/API.",
		}}
	}

	updatedChannel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName,
		&packermodels.HashicorpCloudPackerIteration{
			IncrementalVersion: int32(d.Get("iteration_version").(int)),
			ID:                 d.Get("iteration_id").(string),
			Fingerprint:        d.Get("iteration_fingerprint").(string),
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
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}
	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	updatedChannel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName,
		&packermodels.HashicorpCloudPackerIteration{
			IncrementalVersion: int32(d.Get("iteration_version").(int)),
			ID:                 d.Get("iteration_id").(string),
			Fingerprint:        d.Get("iteration_fingerprint").(string),
		}, nil,
	)
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
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("channel_name").(string)

	_, err = clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName,
		&packermodels.HashicorpCloudPackerIteration{
			IncrementalVersion: 0,
			ID:                 "",
			Fingerprint:        "",
		}, nil,
	)
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
	if err := setLocationData(d, loc); err != nil {
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

	var iteration *packermodels.HashicorpCloudPackerIteration
	if id, ok := d.GetOk("iteration_id"); ok && d.HasChange("iteration_id") && id.(string) != "" {
		iteration, err = clients.GetIterationFromID(ctx, client, loc, bucketName, id.(string))
	} else if fingerprint, ok := d.GetOk("iteration_fingerprint"); ok && d.HasChange("fingerprint_id") && fingerprint.(string) != "" {
		iteration, err = clients.GetIterationFromFingerprint(ctx, client, loc, bucketName, fingerprint.(string))
	} else if version, ok := d.GetOk("iteration_version"); ok && d.HasChange("iteration_version") && version.(int) > 0 {
		iteration, err = clients.GetIterationFromVersion(ctx, client, loc, bucketName, int32(version.(int)))
	}
	if err != nil {
		return err
	}

	if iteration == nil {
		iteration = &packermodels.HashicorpCloudPackerIteration{
			ID:                 "",
			Fingerprint:        "",
			IncrementalVersion: 0,
		}
	}

	if err := d.SetNew("iteration_id", iteration.ID); err != nil {
		return err
	}

	if err := d.SetNew("iteration_fingerprint", iteration.Fingerprint); err != nil {
		return err
	}

	if err := d.SetNew("iteration_version", iteration.IncrementalVersion); err != nil {
		return err
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

	if err := d.Set("iteration_id", iteration.ID); err != nil {
		return err
	}

	if err := d.Set("iteration_fingerprint", iteration.Fingerprint); err != nil {
		return err
	}

	return nil
}
