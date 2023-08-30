// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"golang.org/x/exp/slices"
)

var defaultPackerTimeout = time.Minute

func dataSourcePackerImage() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Image data source gets an image associated with an iteration, either from a specific iteration or from the iteration assigned to a given Channel.",
		ReadContext: dataSourcePackerImageRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPackerTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"bucket_name": {
				Description:      "The slug of the HCP Packer Registry bucket where this image is located.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"cloud_provider": {
				Description: "Name of the cloud provider where this image is stored.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"region": {
				Description: "The Region where this image is stored, if any.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the HCP Packer Registry image is located.
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IsUUID,
			},
			"iteration_id": {
				Description:  "The iteration from which to get the image. Either this or `channel` must be specified.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"iteration_id", "channel"},
			},
			"channel": {
				Description:  "The channel that points to the version of the image being retrieved. Either this or `iteration_id` must be specified. Note: will incur a billable request",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"iteration_id", "channel"},
			},
			"component_type": {
				Description: "Name of the builder that built this image. Ex: `amazon-ebs.example`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			// computed outputs
			"organization_id": {
				Description: "The ID of the organization where this image is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of this image.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"build_id": {
				Description: "HCP ID of the build containing this image.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_image_id": {
				Description: "Cloud Image ID or URL string identifying this image for the builder that built it.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"packer_run_uuid": {
				Description: "UUID of the build containing this image.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"labels": {
				Description: "Labels associated with the build containing this image.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			"revoke_at": {
				Description: "The revocation time of the iteration containing this image. This field will be null for any iteration that has not been revoked or scheduled for revocation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePackerImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket_name").(string)
	channelSlug := d.Get("channel").(string)
	iterationID := d.Get("iteration_id").(string)
	componentType := d.Get("component_type").(string)
	cloudProvider := d.Get("cloud_provider").(string)
	region := d.Get("region").(string)

	log.Printf("[INFO] Reading HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s/iteration_id=%s]", bucketName, loc.ProjectID, loc.OrganizationID, channelSlug, iterationID)

	var iteration *models.HashicorpCloudPackerIteration

	if iterationID != "" {
		iteration, err = clients.GetIterationFromID(
			ctx,
			client,
			loc,
			bucketName,
			iterationID,
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if channelSlug != "" {
		channel, err := clients.GetPackerChannelBySlug(
			ctx,
			client,
			loc,
			bucketName,
			channelSlug,
		)
		if err != nil {
			return diag.FromErr(err)
		}
		if channel.Iteration == nil {
			return diag.Errorf("Channel does not have an assigned iteration (channel: %s)", channelSlug)
		}
		iteration = channel.Iteration
	}

	var build *models.HashicorpCloudPackerBuild
	var image *models.HashicorpCloudPackerImage
	for _, b := range iteration.Builds {
		if b.CloudProvider != cloudProvider {
			continue
		}
		if componentType != "" && b.ComponentType != componentType {
			continue
		}
		index := slices.IndexFunc(
			b.Images,
			func(image *models.HashicorpCloudPackerImage) bool {
				return image.Region == region
			},
		)
		if index >= 0 {
			build = b
			image = b.Images[index]
			break
		}
	}
	if build == nil || image == nil {
		return diag.Errorf("Could not find image with attributes (region: %q cloud: %q, iteration: %q, component_type: %q).", region, cloudProvider, iterationID, componentType)
	}

	d.SetId(image.ID)
	if err := d.Set("component_type", build.ComponentType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", image.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("build_id", build.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cloud_image_id", image.ImageID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("iteration_id", iteration.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("packer_run_uuid", build.PackerRunUUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("labels", build.Labels); err != nil {
		return diag.FromErr(err)
	}
	if !iteration.RevokeAt.IsZero() {
		if err := d.Set("revoke_at", iteration.RevokeAt.String()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
