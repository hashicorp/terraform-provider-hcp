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
			"hvn_2": {
				Description: "The unique URL of one of the HVNs being peered.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the peering connection is located. Always matches the HVNs' organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the peering connection is located. Always matches the HVNs' project.",
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
	orgID := client.Config.OrganizationID

	peeringID := d.Get("peering_id").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnLink1, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink1.ID, loc)
	if err != nil {
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	// Peering connection found, update resource data.
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
