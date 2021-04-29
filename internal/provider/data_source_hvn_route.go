package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceHVNRoute() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN Route data source provides information about an existing HVN route.",
		ReadContext: dataSourceHVNRouteRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnRouteDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"destination_cidr": {
				Description: "The destination CIDR of the HVN route",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"self_link": {
				Description: "A unique URL identifying the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"target_link": {
				Description: "A unique URL identifying the target of the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the HVN route was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceHVNRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn := d.Get("hvn").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink

	hvnLink, err := parseLinkURL(hvn, HvnResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	destination := d.Get("destination_cidr").(string)

	log.Printf("[INFO] Reading HVN route for HVN (%s) with destination_cidr=%s ", hvn, destination)
	route, err := clients.ListHVNRoutes(ctx, client, hvnLink.ID, destination, "", "", loc)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN route for HVN (%s) with destination_cidr=%s: %v",
			hvn, destination, err)
	}

	// ListHVNRoutes call should return 1 and only 1 HVN route.
	if len(route) > 1 {
		return diag.Errorf("Unexpected number of HVN routes returned for destination_cidr=%s: %d", destination, len(route))
	}
	if len(route) == 0 {
		return diag.Errorf("No HVN route found for destionation_cidr=%s", destination)
	}

	link := newLink(loc, HVNRouteResourceType, route[0].ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	if err := setHVNRouteResourceData(d, route[0], loc); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
