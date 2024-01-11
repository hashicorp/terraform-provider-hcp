// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"log"

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
			"project_id": {
				Description: "The ID of the HCP project where the HVN route is located. Always matches the project ID in `hvn_link`. Setting this attribute is deprecated, but it will remain usable in read-only form.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Deprecated: `
Setting the 'project_id' attribute is deprecated, but it will remain usable in read-only form.
Previously, the value for this attribute was required to match the project ID contained in 'hvn_link'. Now, the value will be calculated automatically.
Remove this attribute from the configuration for any affected resources.
`,
			},
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
			"azure_config": {
				Description: "The azure configuration for routing.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"next_hop_type": {
							Description: "The type of Azure hop the packet should be sent to.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"next_hop_ip_address": {
							Description: "Contains the IP address packets should be forwarded to. Next hop values are only allowed in routes where the next hop type is VirtualAppliance.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
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
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID := d.Get("hvn_route_id").(string)
	routeLink := newLink(hvnLink.Location, HVNRouteResourceType, routeID)
	routeURL, err := linkURL(routeLink)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(routeURL)

	log.Printf("[INFO] Reading HVN route (%s)", routeID)
	route, err := clients.GetHVNRoute(ctx, client, hvnLink.ID, routeID, hvnLink.Location)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN route (%s): %v", routeID, err)
	}

	// HVN route found, update resource data.
	if err := setHVNRouteResourceData(d, route, hvnLink.Location); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
