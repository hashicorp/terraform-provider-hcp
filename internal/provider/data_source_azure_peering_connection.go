// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceAzurePeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:        "The Azure peering connection data source provides information about a peering connection between an HVN and a peer Azure VNet.",
		ReadWithoutTimeout: dataSourceAzurePeeringConnectionRead,
		Timeouts: &schema.ResourceTimeout{
			Read: &peeringCreateTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"peering_id": {
				Description: "The ID of the peering connection.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hvn_link": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional inputs
			"wait_for_active_state": {
				Description: "If `true`, Terraform will wait for the peering connection to reach an `ACTIVE` state before continuing. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the peering connection is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the peering connection is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"application_id": {
				Description: "The ID of the Azure application whose credentials are used to peer the HCP HVN's underlying VNet with the customer VNet.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vnet_name": {
				Description: "The name of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_subscription_id": {
				Description: "The subscription ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vnet_region": {
				Description: "The region of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_tenant_id": {
				Description: "The tenant ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_resource_group_name": {
				Description: "The resource group name of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"azure_peering_id": {
				Description: "The peering connection ID used by Azure.",
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
				Description: "The state of the Azure peering connection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAzurePeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	orgID := client.Config.OrganizationID

	peeringID := d.Get("peering_id").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}
	waitForActive := d.Get("wait_for_active_state").(bool)

	// Query for the peering.
	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink.ID, loc)
	if err != nil {
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	// Set the globally unique id of this peering in state.
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Store resource data.
	if err := setAzurePeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	// Skip waiting.
	if !waitForActive || *peering.State == models.HashicorpCloudNetwork20200907PeeringStateACTIVE {
		return nil
	}

	// If it's not in a state where it could later become ACTIVE, we're going to bail.
	terminalState := true
	for _, state := range clients.WaitForPeeringToBeActiveStates {
		if state == string(*peering.State) {
			terminalState = false
			break
		}
	}

	// If it's not in a state that we should wait on, issue a warning and bail.
	if terminalState {
		return []diag.Diagnostic{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Peering is in an unexpected state, connections may fail: %q", string(*peering.State)),
			Detail:   "Expected a CREATING, PENDING_ACCEPTANCE, ACCEPTED, or ACTIVE state",
		}}
	}

	// Store resource data again, updating Peering state.
	var result []diag.Diagnostic
	peering, err = clients.WaitForPeeringToBeActive(ctx, client, peering.ID, hvnLink.ID, loc, peeringCreateTimeout)
	if peering != nil {
		if err := setAzurePeeringResourceData(d, peering); err != nil {
			result = diag.FromErr(err)
		}
	}

	// If we didn't reach the desired state, throw a diagnostic err.
	if err != nil {
		for _, d := range diag.FromErr(err) {
			result = append(result, d)
		}
	}
	return result
}
