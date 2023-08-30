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

func dataSourceHvnPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN peering connection data source provides information about an existing peering connection between HVNs.",
		ReadContext: dataSourceHvnPeeringConnectionRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"peering_id": {
				Description: "The ID of the peering connection.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hvn_1": {
				Description: "The unique URL of one of the HVNs being peered.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"hvn_2": {
				Description: "The unique URL of one of the HVNs being peered. Setting this attribute is deprecated, but it will remain usable in read-only form.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Deprecated: `
Setting the 'hvn_2' attribute is deprecated, but it will remain usable in read-only form.
Previously, the value for this attribute was not used to fetch data, and it was not validated against the actual value of 'hvn_2'. Now, the value will be populated automatically.
Remove this attribute from the configuration for any affected resources.
`,
			},
			"project_id": {
				Description: "The ID of the HCP project where the HVN peering connection is located. Always matches hvn_1's project ID. Setting this attribute is deprecated, but it will remain usable in read-only form.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Deprecated: `
Setting the 'project_id' attribute is deprecated, but it will remain usable in read-only form.
Previously, the value for this attribute was required to match the project ID contained in 'hvn_1'. Now, the value will be calculated automatically.
Remove this attribute from the configuration for any affected resources.
`,
			},
			"organization_id": {
				Description: "The ID of the HCP organization where the peering connection is located. Always matches both HVNs' organization ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the peering connection was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the peering connection will be considered expired if it hasn't transitioned into `ACCEPTED` or `ACTIVE` state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the peering connection",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the HVN peering connection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceHvnPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn1Link, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := d.Get("peering_id").(string)
	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvn1Link.ID, hvn1Link.Location)
	if err != nil {
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	// Peering connection found, update resource data.
	hvn2Link := newLink(peering.Target.HvnTarget.Hvn.Location, HvnResourceType, peering.Target.HvnTarget.Hvn.ID)
	hvn2URL, err := linkURL(hvn2Link)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hvn_2", hvn2URL); err != nil {
		return diag.FromErr(err)
	}

	if err := setHvnPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	// Set the globally unique id of this peering in the state.
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	return nil
}
