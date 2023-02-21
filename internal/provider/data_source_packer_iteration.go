// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourcePackerIteration() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Image data source iteration gets the most recent iteration (or build) of an image, given a channel.",
		ReadContext: dataSourcePackerIterationRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPackerTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"bucket_name": {
				Description:      "The slug of the HCP Packer Registry image bucket to pull from.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"channel": {
				Description:      "The channel that points to the version of the image you want.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// computed outputs
			"author_id": {
				Description: "The name of the person who created this iteration.",
				Type:        schema.TypeString,
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
			"fingerprint": {
				Description: "The unique fingerprint associated with this iteration; often a git sha.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ulid": {
				Description: "The ULID of this iteration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// Actual iteration:
			"incremental_version": {
				Description: "Incremental version of this iteration",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of this iteration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "Time this build was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"revoke_at": {
				Description: "The revocation time of this iteration. This field will be null for any iteration that has not been revoked or scheduled for revocation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePackerIterationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	channelSlug := d.Get("channel").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s]", bucketName, loc.ProjectID, loc.OrganizationID, channelSlug)

	channel, err := clients.GetPackerChannelBySlug(ctx, client, loc, bucketName, channelSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	if channel.Iteration == nil {
		return diag.Errorf("no iteration information found for the specified channel %s", channelSlug)
	}

	iteration := channel.Iteration

	d.SetId(iteration.ID)

	if err := d.Set("author_id", iteration.AuthorID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bucket_name", iteration.BucketSlug); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", iteration.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fingerprint", iteration.Fingerprint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ulid", iteration.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("incremental_version", iteration.IncrementalVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", iteration.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if !time.Time(iteration.RevokeAt).IsZero() {
		if err := d.Set("revoke_at", iteration.RevokeAt.String()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
