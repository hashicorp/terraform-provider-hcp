package provider

import (
	"context"
	"log"
	"time"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// a minute sounds like a lot already since this would be mainly pulling data.
var defaultPackerTimeout = time.Minute

func dataSourcePackerImageIteration() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Image data source iteration gets the most recent iteration (or build) of an image given an channel.",
		ReadContext: dataSourcePackerImageRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPackerTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"bucket": {
				Description:      "The slug of the HCP Packer Registry image bucket to pull from.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"channel": {
				Description:      "The channel that points to the version of the image you want.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSlugID,
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

			// Actual iteration:

			"id": {
				Description: "ID of this iteration",
				Type:        schema.TypeString,
				Computed:    true,
			},
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
			"builds": {
				Description: "Builds for this iteration",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cloud_provider": {
								Type: schema.TypeString,
							},
							"component_type": {
								Type: schema.TypeString,
							},
							"created_at": {
								Type: schema.TypeString,
							},
							// "id": {
							// 	Type: schema.TypeString,
							// },
							"images": {
								Type: schema.TypeList,
								Elem: &schema.Schema{
									Type: schema.TypeSet,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"created_at": {
												Type: schema.TypeString,
											},
											"id": {
												Type: schema.TypeString,
											},
											"image_id": {
												Type: schema.TypeString,
											},
											"region": {
												Type: schema.TypeString,
											},
										},
									},
								},
							},
							"iteration_id": {
								Type: schema.TypeString,
							},
							"labels": {
								Type: schema.TypeMap,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"packer_run_uuid": {
								Type: schema.TypeString,
							},
							"status": {
								Type: schema.TypeString,
							},
							"updated_at": {
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourcePackerImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("bucket").(string)
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

	iteration := channel.Pointer.Iteration

	d.SetId(channel.Pointer.Iteration.ID)

	d.Set("incremental_version", iteration.IncrementalVersion)
	d.Set("created_at", iteration.CreatedAt.String())
	d.Set("builds", flattenPackerBuildList(iteration.Builds))

	return nil
}

func setLocationData(d *schema.ResourceData, loc *sharedmodels.HashicorpCloudLocationLocation) error {
	if err := d.Set("organization_id", loc.OrganizationID); err != nil {
		return err
	}

	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}
	return nil
}

func flattenPackerBuildList(builds []*packermodels.HashicorpCloudPackerBuild) (flattened []map[string]interface{}) {
	for _, build := range builds {
		out := map[string]interface{}{
			"cloud_provider":  build.CloudProvider,
			"component_type":  build.ComponentType,
			"created_at":      build.CreatedAt.String(),
			"id":              build.ID,
			"images":          flattenPackerBuildImagesList(build.Images),
			"iteration_id":    build.IterationID,
			"labels":          build.Labels,
			"packer_run_uuid": build.PackerRunUUID,
			"status":          build.Status,
			"updated_at":      build.UpdatedAt.String(),
		}
		flattened = append(flattened, out)
	}
	return
}

func flattenPackerBuildImagesList(images []*packermodels.HashicorpCloudPackerImage) (flattened []map[string]interface{}) {
	for _, image := range images {
		out := map[string]interface{}{
			"created_at": image.CreatedAt.String(),
			"id":         image.ID,
			"image_id":   image.ImageID,
			"region":     image.Region,
		}
		flattened = append(flattened, out)
	}
	return
}
