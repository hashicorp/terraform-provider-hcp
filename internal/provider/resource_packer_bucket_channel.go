package provider

import (
	"context"
	"errors"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

//resource hcp_packer_channel "example" {
//	 bucket_name = ""
//description
//assigned_iteration_id

//}
// output {
//id = channel-id
//description = "channel description"
//assigned_iteration_id = "iteration id"
//organization_id
//project_id
//created_at
//
//}

func resourcePackerBucketChannel() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Bucket Channel resource allow you to manage a channel within an active HCP Packer Registry bucket.",
		CreateContext: resourcePackerBucketChannelCreate,
		DeleteContext: resourcePackerBucketChannelDelete,
		ReadContext:   resourcePackerBucketChannelRead,
		UpdateContext: resourcePackerBucketChannelUpdate,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultPackerTimeout,
			Default: &defaultPackerTimeout,
			Update:  &defaultPackerTimeout,
			Delete:  &defaultPackerTimeout,
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Description:      "The slug of the HCP Packer Registry image bucket where the channel should be managed in.",
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional inputs
			"assigned_iteration_id": {
				Description: "The iteration id to assign to the channel.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			// computed outputs
			"assigned_iteration_version": {
				Description: "The incremental version of the iteration assigned to the channel.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"organization_id": {
				Description: "The ID of the organization this HCP Packer registry is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Packer registry is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of this build.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePackerBucketChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)

	}
	// TODO (nywilken) may need to look into pagination depending on channel list sizes
	resp, err := clients.ListBucketChannels(ctx, client, loc, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}

	channelName := d.Get("name").(string)
	if len(resp.Channels) == 0 {
		return diag.Errorf("Unable to find channel in bucket %s named %s.", bucketName, channelName)
	}

	// iteration to find channel
	var channel packermodels.HashicorpCloudPackerChannel
	for _, c := range resp.Channels {
		// do we have state?
		if c.Slug == channelName {
			channel = *c
		}
	}

	return setPackerBucketChannelResourceData(d, &channel)
}

func resourcePackerBucketChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	client := meta.(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	channelInput := clients.ChannelResourceInput{
		Name:                channelName,
		BucketName:          bucketName,
		AssignedIterationID: d.Get("assigned_iteration_id").(string),
	}

	channel, err := clients.CreateBucketChannel(ctx, client, loc, channelInput)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel == nil {
		return diag.Errorf("Unable to find channel in bucket %s named %s.", bucketName, channelName)
	}

	return setPackerBucketChannelResourceData(d, channel)
}

func resourcePackerBucketChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	channelInput := clients.ChannelResourceInput{
		BucketName:          d.Get("bucket_name").(string),
		Name:                d.Get("name").(string),
		AssignedIterationID: d.Get("assigned_iteration_id").(string),
	}

	client := meta.(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	channel, err := clients.UpdateBucketChannel(ctx, client, loc, channelInput)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPackerBucketChannelResourceData(d, channel)
}

func resourcePackerBucketChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelName := d.Get("name").(string)

	client := meta.(*clients.Client)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	_, err := clients.DeleteBucketChannel(ctx, client, loc, bucketName, channelName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setPackerBucketChannelResourceData(d *schema.ResourceData, channel *packermodels.HashicorpCloudPackerChannel) diag.Diagnostics {

	if channel == nil {
		err := errors.New("unexpected empty provided when setting state")
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)
	if err := d.Set("created_at", channel.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if channel.Iteration == nil {
		return nil
	}

	if err := d.Set("assigned_iteration_id", channel.Iteration.ID); err != nil {
		return diag.FromErr(err)

	}
	if err := d.Set("assigned_iteration_version", channel.Iteration.IncrementalVersion); err != nil {
		return diag.FromErr(err)

	}

	return nil
}
