// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"time"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var defaultPackerTimeout = time.Minute

func dataSourcePackerImage() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Image data source iteration gets the most recent iteration (or build) of an image, given an iteration id or a channel.",
		ReadContext: dataSourcePackerImageRead,
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
			"cloud_provider": {
				Description: "Name of the cloud provider this image is stored-in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"region": {
				Description: "Region this image is stored in, if any.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional inputs
			"iteration_id": {
				Description:  "The iteration from which to get the image. Either this or `channel` must be specified.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"channel"},
			},
			"channel": {
				Description: "The channel that points to the version of the image being retrieved. Either this or `iteration_id` must be specified. Note: will incur a billable request",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"component_type": {
				Description: "Name of the builder that built this image. Ex: `amazon-ebs.example`.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			// computed outputs
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
			"build_id": {
				Description: "HCP ID of this build.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_image_id": {
				Description: "Cloud Image ID or URL string identifying this image for the builder that built it.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"packer_run_uuid": {
				Description: "UUID of this build.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"labels": {
				Description: "Labels associated with this build.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			"revoke_at": {
				Description: "The revocation time of this build. This field will be null for any build that has not been revoked or scheduled for revocation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePackerImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket_name").(string)
	cloudProvider := d.Get("cloud_provider").(string)
	region := d.Get("region").(string)
	channelSlug := d.Get("channel").(string)
	componentType := d.Get("component_type").(string)
	iterationID := d.Get("iteration_id").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	if err := setLocationData(d, loc); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading HCP Packer registry (%s) [project_id=%s, organization_id=%s, channel=%s/iteration_id=%s]", bucketName, loc.ProjectID, loc.OrganizationID, channelSlug, iterationID)

	var iteration *packermodels.HashicorpCloudPackerIteration
	var err error

	if iterationID != "" {
		iteration, err = clients.GetIterationFromID(
			ctx,
			client,
			loc,
			bucketName,
			iterationID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	var channel *packermodels.HashicorpCloudPackerChannel

	if channelSlug != "" {
		channel, err = clients.GetPackerChannelBySlug(
			ctx,
			client,
			loc,
			bucketName,
			channelSlug)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Assuming we passed the above check, the rest of the channel is not
	// used after that,
	if channel != nil {
		iteration = channel.Iteration
	}

	found := false
	for _, build := range iteration.Builds {
		if build.CloudProvider != cloudProvider {
			continue
		}
		for _, image := range build.Images {
			if image.Region == region && filterBuildByComponentType(build, componentType) {
				found = true
				d.SetId(image.ID)
				if err := d.Set("component_type", build.ComponentType); err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("created_at", build.CreatedAt.String()); err != nil {
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
				if !time.Time(iteration.RevokeAt).IsZero() {
					if err := d.Set("revoke_at", iteration.RevokeAt.String()); err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	if !found {
		return diag.Errorf("Unable to load image (region: %s, cloud: %s, iteration: %s, component_type: %s).", region, cloudProvider, iterationID, componentType)
	}

	return nil
}

func filterBuildByComponentType(build *packermodels.HashicorpCloudPackerBuild, componentType string) bool {
	// optional field is not specified, passthrough
	if componentType == "" {
		return true
	}
	// if specified, only the matched image metadata is returned by this effect
	return build.ComponentType == componentType
}
