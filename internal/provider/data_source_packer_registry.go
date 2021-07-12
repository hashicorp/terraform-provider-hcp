package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourcePackerRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Registry data source provides information about an existing image build stored in the Packer registry",
		ReadContext: dataSourcePackerRegistryRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPackerRegistryTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"bucket_id": {
				Description:      "The ID of the HCP Packer Registry image bucket to pull from.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"channel": {
				Description:      "The channel that points to the version of the image you want.",
				Type:             schema.TypeString,
				Required:         false,
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
			"incremental_version": {
				Description: "The Packer version of the registry.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the Packer registry was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePackerRegistryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketID := d.Get("bucket_id").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	log.Printf("[INFO] Reading HCP Packer registry (%s) [project_id=%s, organization_id=%s]", registryID, loc.ProjectID, loc.OrganizationID)

	bucket, err := clients.GetPackerBucketByID(ctx, client, loc, registryID)
	if err != nil {
		return diag.FromErr(err)
	}

	// build the id for this Packer registry
	link := newLink(loc, PackerRegistryResourceType, registryID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// registry found, update resource data.
	if err := setPackerRegistryResourceData(d, registry); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
