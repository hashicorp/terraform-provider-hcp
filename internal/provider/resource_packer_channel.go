// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"google.golang.org/grpc/codes"
)

func resourcePackerChannel() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Channel resource allows you to manage a bucket channel within an active HCP Packer Registry.",
		CreateContext: resourcePackerChannelCreate,
		DeleteContext: resourcePackerChannelDelete,
		ReadContext:   resourcePackerChannelRead,
		UpdateContext: resourcePackerChannelUpdate,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultPackerTimeout,
			Default: &defaultPackerTimeout,
			Update:  &defaultPackerTimeout,
			Delete:  &defaultPackerTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourcePackerChannelImport,
		},
		CustomizeDiff: resourcePackerChannelCustomizeDiff,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"name": {
				Description:      "The name of the channel being managed.",
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"bucket_name": {
				Description:      "The slug of the HCP Packer Registry bucket where the channel should be created.",
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional inputs
			"restricted": {
				Description: "If true, the channel is only visible to users with permission to create and manage it. Otherwise the channel is visible to every member of the organization.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"project_id": {
				Description: `
The ID of the HCP project where this channel is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Computed:     true,
			},
			// Computed Values
			"author_id": {
				Description: "The author of this channel.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The creation time of this channel.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization_id": {
				Description: "The ID of the HCP organization where this channel is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The time this channel was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"managed": {
				Description: "If true, the channel is an HCP Packer managed channel",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourcePackerChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	log.Printf("[INFO] Reading HCP Packer channel (%s) [bucket_name=%s, project_id=%s, organization_id=%s]", channelName, bucketName, loc.ProjectID, loc.OrganizationID)

	channel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel == nil {
		log.Printf(
			"[WARN] HCP Packer channel with (name %q) (bucket_name %q) (project_id %q) not found, removing from state.",
			channelName, bucketName, loc.ProjectID,
		)
		d.SetId("")
		return nil
	}
	return setPackerChannelResourceData(d, channel)
}

func resourcePackerChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	createRestriction := packermodels.NewHashicorpCloudPackerCreateChannelRequestRestriction(packermodels.HashicorpCloudPackerCreateChannelRequestRestrictionRESTRICTIONUNSET)
	//lint:ignore SA1019 GetOkExists is fine for use with booleans and has defined behavior
	restrictedRaw, restrictedSet := d.GetOkExists("restricted")
	if restrictedSet {
		if restrictedRaw.(bool) {
			createRestriction = packermodels.NewHashicorpCloudPackerCreateChannelRequestRestriction(packermodels.HashicorpCloudPackerCreateChannelRequestRestrictionRESTRICTED)
		} else {
			createRestriction = packermodels.NewHashicorpCloudPackerCreateChannelRequestRestriction(packermodels.HashicorpCloudPackerCreateChannelRequestRestrictionUNRESTRICTED)
		}
	}

	newChannel, err := clients.CreatePackerChannel(ctx, client, loc, bucketName, channelName, createRestriction)
	if err == nil {
		if newChannel == nil {
			return diag.Errorf("expected a non-nil channel from CreateChannel, but got nil")
		}

		// Handle successfully created channel
		return setPackerChannelResourceData(d, newChannel)
	}

	// Check error to see if the channel is a pre-existing managed channel
	errCreate, ok := err.(*packer_service.PackerServiceCreateChannelDefault)
	if !ok {
		return diag.FromErr(err)
	}
	if payload := errCreate.Payload; payload == nil || codes.Code(payload.Code) != codes.AlreadyExists {
		return diag.FromErr(errCreate)
	}

	// Channel already exists
	existingChannel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.Errorf("channel already exists. GetChannel failed unexpectedly: %v", err)
	}
	if existingChannel == nil {
		return diag.Errorf("channel already exists. Expected a non-nil channel from GetChannel, but got nil")
	}
	if !existingChannel.Managed {
		return diag.Errorf("channel already exists, use `terraform import` to add it to the terraform state")
	}

	// Channel is managed, attempt update
	diags := diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "This channel already exists and is managed by HCP Packer, so it cannot be manually created.",
			Detail:   "Attempting to automatically adopt the channel. This action will not create a new channel, as the channel already exists.",
		},
	}
	diags = append(diags, setPackerChannelResourceData(d, existingChannel)...)
	if diags.HasError() {
		return diags
	}

	if restrictedSet {
		updatedChannel, err := clients.UpdatePackerChannel(ctx, client, loc, bucketName, channelName, restrictedRaw.(bool))
		if err != nil {
			diags := append(diags, diag.Errorf("UpdateChannel failed unexpectedly: %v", err)...)
			return diags
		}
		if updatedChannel == nil {
			diags := append(diags, diag.Errorf("Expected non-nil channel from UpdateChannel, but got nil")...)
			return diags
		}
		diags = append(diags, setPackerChannelResourceData(d, updatedChannel)...)
	}

	return diags
}

func resourcePackerChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	//lint:ignore SA1019 GetOkExists is fine for use with booleans and has defined behavior
	restrictedRaw, ok := d.GetOkExists("restricted")
	if !ok {
		// Currently only the restriction can be updated, and it should not be
		// updated if it isn't set in the config, so we return.
		// This should never happen because all other fields are ForceNew
		return nil
	}

	channel, err := clients.UpdatePackerChannel(ctx, client, loc, bucketName, channelName, restrictedRaw.(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	return setPackerChannelResourceData(d, channel)
}

func resourcePackerChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	if d.Get("managed").(bool) {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "This channel is managed by HCP Packer, so it cannot be deleted.",
			Detail:   "The channel has been removed from the terraform state, but has not been deleted from HCP Packer.",
		}}
	}

	_, err = clients.DeletePackerChannel(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackerChannelImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_packer_channel.test {project_id}:{bucket_name}:{channel_name}
	// use default project ID from provider:
	//   terraform import hcp_packer_channel.test {bucket_name}:{channel_name}

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

	channel, err := clients.GetPackerChannelBySlugFromList(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, fmt.Errorf("unable to find channel in bucket %s named %s", bucketName, channelName)
	}

	d.SetId(channel.ID)
	if err := d.Set("bucket_name", bucketName); err != nil {
		return nil, err
	}
	if err := d.Set("name", channelName); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePackerChannelCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if d.HasChanges("restricted") {
		if err := d.SetNewComputed("updated_at"); err != nil {
			return err
		}
		if err := d.SetNewComputed("author_id"); err != nil {
			return err
		}
	}

	return nil
}

func setPackerChannelResourceData(d *schema.ResourceData, channel *packermodels.HashicorpCloudPackerChannel) diag.Diagnostics {
	if channel == nil {
		err := errors.New("unexpected empty channel provided when setting state")
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)

	if err := d.Set("author_id", channel.AuthorID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", channel.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("updated_at", channel.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("restricted", channel.Restricted); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("managed", channel.Managed); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
