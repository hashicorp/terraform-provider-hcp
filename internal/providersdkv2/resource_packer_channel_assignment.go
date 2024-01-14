// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

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
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv1"
)

// This string is used as the version fingerprint to represent an unassigned
// (or "null") channel assignment
const unassignString string = "none"

func resourcePackerChannelAssignment() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Channel Assignment resource allows you to manage the version assigned to a channel in an active HCP Packer Registry.",
		CreateContext: resourcePackerChannelAssignmentCreate,
		DeleteContext: resourcePackerChannelAssignmentDelete,
		ReadContext:   resourcePackerChannelAssignmentRead,
		UpdateContext: resourcePackerChannelAssignmentUpdate,
		CustomizeDiff: resourcePackerChannelAssignmentCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultPackerTimeout,
			Default: &defaultPackerTimeout,
			Update:  &defaultPackerTimeout,
			Delete:  &defaultPackerTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourcePackerChannelAssignmentImport,
		},
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
				Description:      "The slug of the HCP Packer bucket where the channel is located.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"iteration_fingerprint": {
				Description:  "The fingerprint of the version assigned to the channel.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_fingerprint", "version_fingerprint"},
				ValidateFunc: validation.StringIsNotEmpty,
				Deprecated: "This attribute has been renamed to 'version_fingerprint`. The `iteration_fingerprint` attribute is deprecated and will be removed in a future release. " +
					"The `iteration_fingerprint` attribute will act as an alias for `version_fingerprint` during the deprecation period.",
			},
			"version_fingerprint": {
				Description:  "The fingerprint of the version assigned to the channel.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_fingerprint", "version_fingerprint"},
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"project_id": {
				Description: `
The ID of the HCP project where the channel is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
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

	channel, err := packerv1.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel == nil || channel.Slug == "" {
		d.SetId("")
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found.", channelName, bucketName, loc.ProjectID),
		}}
	}

	if err := setPackerChannelAssignmentVersionData(d, channel.Iteration); err != nil {
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

	channel, err := packerv1.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	} else if channel == nil || channel.Slug == "" {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found.", channelName, bucketName, loc.ProjectID),
		}}
	} else if channel.Managed {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) is managed by HCP Packer and cannot have a version assigned by Terraform.", channelName, bucketName, loc.ProjectID),
		}}
	} else if version := channel.Iteration; version != nil && (version.IncrementalVersion > 0 || version.ID != "" || version.Fingerprint != "") {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) already has an assigned version.", channelName, bucketName, loc.ProjectID),
			Detail:   "To adopt this resource into Terraform, use `terraform import`, or remove the channel's assigned version using the HCP Packer GUI/API.",
		}}
	}

	versionFingerprint := d.Get("version_fingerprint").(string)
	if versionFingerprint == "" {
		versionFingerprint = d.Get("iteration_fingerprint").(string)
	}
	if versionFingerprint == unassignString {
		versionFingerprint = ""
	}

	updatedChannel, err := packerv1.UpdatePackerChannelAssignment(ctx, client, loc, bucketName, channelName, versionFingerprint)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(updatedChannel.ID)

	if err := setPackerChannelAssignmentVersionData(d, updatedChannel.Iteration); err != nil {
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

	versionFingerprint := d.Get("version_fingerprint").(string)
	if versionFingerprint == "" {
		versionFingerprint = d.Get("iteration_fingerprint").(string)
	}
	if versionFingerprint == unassignString {
		versionFingerprint = ""
	}

	updatedChannel, err := packerv1.UpdatePackerChannelAssignment(ctx, client, loc, bucketName, channelName, versionFingerprint)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setPackerChannelAssignmentVersionData(d, updatedChannel.Iteration); err != nil {
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

	_, err = packerv1.UpdatePackerChannelAssignment(ctx, client, loc, bucketName, channelName, "")
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

	channel, err := packerv1.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return nil, err
	} else if channel == nil || channel.Slug == "" {
		return nil, fmt.Errorf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) not found", channelName, bucketName, loc.ProjectID)
	} else if channel.Managed {
		return nil, fmt.Errorf("HCP Packer channel with (channel_name %q) (bucket_name %q) (project_id %q) is managed by HCP Packer and cannot have a version assigned by Terraform", channelName, bucketName, loc.ProjectID)
	}

	d.SetId(channel.ID)

	if err := setPackerChannelAssignmentVersionData(d, channel.Iteration); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func setPackerChannelAssignmentVersionData(d *schema.ResourceData, i *packermodels.HashicorpCloudPackerIteration) error {
	var version packermodels.HashicorpCloudPackerIteration

	if i == nil {
		version = packermodels.HashicorpCloudPackerIteration{
			Fingerprint: "",
		}
	} else {
		version = *i
	}

	fingerprint := version.Fingerprint
	if fingerprint == "" {
		fingerprint = unassignString
	}
	if err := d.Set("version_fingerprint", fingerprint); err != nil {
		return err
	}
	if err := d.Set("iteration_fingerprint", fingerprint); err != nil {
		return err
	}

	return nil
}

func resourcePackerChannelAssignmentCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if rawFingerprint, ok := d.GetOk("version_fingerprint"); ok && d.HasChange("version_fingerprint") && d.NewValueKnown("version_fingerprint") {
		if err := d.SetNew("iteration_fingerprint", rawFingerprint.(string)); err != nil {
			return err
		}
	} else if rawFingerprint, ok := d.GetOk("iteration_fingerprint"); ok && d.HasChange("iteration_fingerprint") && d.NewValueKnown("iteration_fingerprint") {
		if err := d.SetNew("version_fingerprint", rawFingerprint.(string)); err != nil {
			return err
		}
	}

	return nil
}
