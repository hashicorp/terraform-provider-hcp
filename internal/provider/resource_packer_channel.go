// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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

func resourcePackerChannel() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Channel resource allows you to manage image bucket channels within an active HCP Packer Registry.",
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
				Description:      "The slug of the HCP Packer Registry image bucket where the channel should be created in.",
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where this channel is located in. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Computed:     true,
			},
			"iteration": {
				Description: "The iteration assigned to the channel.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fingerprint": {
							Description:  "The fingerprint of the iteration assigned to the channel.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ExactlyOneOf: []string{"iteration.0.id", "iteration.0.fingerprint", "iteration.0.incremental_version"},
						},
						"id": {
							Description:  "The ID of the iteration assigned to the channel.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ExactlyOneOf: []string{"iteration.0.id", "iteration.0.fingerprint", "iteration.0.incremental_version"},
						},
						"incremental_version": {
							Description:  "The incremental_version of the iteration assigned to the channel.",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ExactlyOneOf: []string{"iteration.0.id", "iteration.0.fingerprint", "iteration.0.incremental_version"},
						},
					},
				},
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
				Description: "The ID of the HCP organization where this channel is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The time this channel was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePackerChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)

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

	channelName := d.Get("name").(string)

	log.Printf("[INFO] Reading HCP Packer channel (%s) [bucket_name=%s, project_id=%s, organization_id=%s]", channelName, bucketName, loc.ProjectID, loc.OrganizationID)

	resp, err := clients.ListBucketChannels(ctx, client, loc, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}

	var channel packermodels.HashicorpCloudPackerChannel
	for _, c := range resp.Channels {
		if c.Slug == channelName {
			channel = *c
			break
		}
	}
	if channel.ID == "" {
		log.Printf(
			"[WARN] HCP Packer chanel with (name %q) (bucket_name %q) (project_id %q) not found, removing from state.",
			channelName, bucketName, loc.ProjectID,
		)
		d.SetId("")
		return nil
	}
	return setPackerChannelResourceData(d, &channel)
}

func resourcePackerChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

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

	iterationConfig, ok := d.GetOk("iteration")
	if !ok {
		channel, err := clients.CreateBucketChannel(ctx, client, loc, bucketName, channelName, nil)
		if err != nil {
			return diag.FromErr(err)
		}

		if channel == nil {
			return diag.Errorf("Unable to create channel in bucket %s named %s.", bucketName, channelName)
		}

		return setPackerChannelResourceData(d, channel)
	}

	var iteration *packermodels.HashicorpCloudPackerIteration
	if config, ok := iterationConfig.([]interface{})[0].(map[string]interface{}); ok {
		iteration = expandIterationConfig(config)
	}

	channel, err := clients.CreateBucketChannel(ctx, client, loc, bucketName, channelName, iteration)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel == nil {
		return diag.Errorf("Unable to create channel in bucket %s named %s.", bucketName, channelName)
	}

	return setPackerChannelResourceData(d, channel)
}

func resourcePackerChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

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

	var iteration *packermodels.HashicorpCloudPackerIteration
	iterationConfig, ok := d.GetOk("iteration")
	if !ok || iterationConfig.([]interface{})[0] == nil {
		channel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName, iteration)
		if err != nil {
			return diag.FromErr(err)
		}
		return setPackerChannelResourceData(d, channel)
	}

	config, ok := iterationConfig.([]interface{})[0].(map[string]interface{})
	if !ok {
		return diag.Errorf("Failed to read iteration configuration during update.")
	}

	updatedIterationConfig := make(map[string]interface{})
	for key, value := range config {
		fullKey := fmt.Sprintf("iteration.0.%s", key)
		// Upstream API doesn't know how to handle the case when all params are set;
		// So we keep the inputs that are not coming from state.
		if d.HasChange(fullKey) {
			updatedIterationConfig[key] = value
		}
	}

	if len(updatedIterationConfig) != 0 {
		iteration = expandIterationConfig(updatedIterationConfig)
	}

	channel, err := clients.UpdateBucketChannel(ctx, client, loc, bucketName, channelName, iteration)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPackerChannelResourceData(d, channel)
}

func resourcePackerChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

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

	_, err = clients.DeleteBucketChannel(ctx, client, loc, bucketName, channelName)
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
	if err := setLocationData(d, loc); err != nil {
		return nil, err

	}
	resp, err := clients.ListBucketChannels(ctx, client, loc, bucketName)
	if err != nil {
		return nil, err
	}

	var channel packermodels.HashicorpCloudPackerChannel
	for _, c := range resp.Channels {
		if c.Slug == channelName {
			channel = *c
			break
		}
	}

	if channel.ID == "" {
		return nil, fmt.Errorf("unable to find channel in bucket %s named %s", bucketName, channelName)
	}

	if channel.Managed {
		return nil, fmt.Errorf("the channel %q is managed by HCP Packer and can not be imported", channel.Slug)
	}

	d.SetId(channel.ID)
	if err := d.Set("bucket_name", bucketName); err != nil {
		return nil, err
	}
	if err := d.Set("name", channelName); err != nil {
		return nil, err
	}

	if channel.Iteration == nil {
		return []*schema.ResourceData{d}, nil
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePackerChannelCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	client := meta.(*clients.Client)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return fmt.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	bucketNameRaw, ok := d.GetOk("bucket_name")
	if !ok {
		return fmt.Errorf("unable to retrieve bucket_name")
	}
	bucketName := bucketNameRaw.(string)

	if d.HasChange("iteration.0") {
		var iterationResponse *packermodels.HashicorpCloudPackerIteration
		var err error
		if id, ok := d.GetOk("iteration.0.id"); d.HasChange("iteration.0.id") && ok && id.(string) != "" {
			iterationResponse, err = clients.GetIterationFromID(ctx, client, loc, bucketName, id.(string))
		} else if fingerprint, ok := d.GetOk("iteration.0.fingerprint"); d.HasChange("iteration.0.fingerprint") && ok && fingerprint.(string) != "" {
			iterationResponse, err = clients.GetIterationFromFingerprint(ctx, client, loc, bucketName, fingerprint.(string))
		} else if version, ok := d.GetOk("iteration.0.incremental_version"); d.HasChange("iteration.0.incremental_version") && ok && version.(int) > 0 {
			iterationResponse, err = clients.GetIterationFromVersion(ctx, client, loc, bucketName, int32(version.(int)))
		}
		if err != nil {
			return err
		}

		iteration := []map[string]interface{}{}
		if iterationResponse != nil {
			iteration = append(iteration, map[string]interface{}{
				"id":                  iterationResponse.ID,
				"fingerprint":         iterationResponse.Fingerprint,
				"incremental_version": iterationResponse.IncrementalVersion,
			})
		}
		err = d.SetNew("iteration", iteration)
		if err != nil {
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

	if err := d.Set("iteration", flattenIterationConfig(channel.Iteration)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("updated_at", channel.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func expandIterationConfig(config map[string]interface{}) *packermodels.HashicorpCloudPackerIteration {
	if config == nil {
		return nil
	}

	var iteration packermodels.HashicorpCloudPackerIteration
	if v, ok := config["id"]; ok && v.(string) != "" {
		iteration.ID = v.(string)
	}
	if v, ok := config["fingerprint"]; ok && v.(string) != "" {
		iteration.Fingerprint = v.(string)
	}
	if v, ok := config["incremental_version"]; ok && v.(int) != 0 {
		iteration.IncrementalVersion = int32(v.(int))
	}

	return &iteration
}

func flattenIterationConfig(iteration *packermodels.HashicorpCloudPackerIteration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if iteration == nil {
		return result
	}

	item := make(map[string]interface{})
	item["id"] = iteration.ID
	item["fingerprint"] = iteration.Fingerprint
	item["incremental_version"] = iteration.IncrementalVersion
	return append(result, item)
}
