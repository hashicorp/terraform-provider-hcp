// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		Description: "The HVN route data source provides information about an existing HVN route.",
		ReadContext: dataSourceHVNRouteRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnRouteDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_link": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hvn_route_id": {
				Description: "The ID of the HVN route.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"self_link": {
				Description: "A unique URL identifying the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"destination_cidr": {
				Description: "The destination CIDR of the HVN route.",
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

	hvn := d.Get("hvn_link").(string)
	hvnLink, err := parseLinkURL(hvn, HvnResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	routeID := d.Get("hvn_route_id").(string)
	routeLink := newLink(loc, HVNRouteResourceType, routeID)
	routeURL, err := linkURL(routeLink)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(routeURL)

	log.Printf("[INFO] Reading HVN route (%s)", routeID)
	route, err := clients.GetHVNRoute(ctx, client, hvnLink.ID, routeID, loc)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN route (%s): %v", routeID, err)
	}

	// HVN route found, update resource data.
	if err := setHVNRouteResourceData(d, route, loc); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
