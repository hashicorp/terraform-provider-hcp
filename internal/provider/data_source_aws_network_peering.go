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

func dataSourceAwsNetworkPeering() *schema.Resource {
	return &schema.Resource{
		Description:        "The AWS network peering data source provides information about an existing network peering between an HVN and a peer AWS VPC.",
		ReadWithoutTimeout: dataSourceAwsNetworkPeeringRead,
		Timeouts: &schema.ResourceTimeout{
			Read: &peeringCreateTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"peering_id": {
				Description:      "The ID of the network peering.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"wait_for_active_state": {
				Description: "If `true`, Terraform will wait for the network peering to reach an `ACTIVE` state before continuing. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the network peering is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the network peering is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_account_id": {
				Description: "The account ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vpc_id": {
				Description: "The ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vpc_region": {
				Description: "The region of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_peering_id": {
				Description: "The peering connection ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the network peering was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the network peering will be considered expired if it hasn't transitioned into `ACCEPTED` or `ACTIVE` state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the network peering.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the network peering.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceAwsNetworkPeeringRead is the func to implement reading of the
// network peering between an HVN and a peer AWS VPC.
func dataSourceAwsNetworkPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnID := d.Get("hvn_id").(string)
	peeringID := d.Get("peering_id").(string)
	waitForActive := d.Get("wait_for_active_state").(bool)

	// Query for the peering.
	log.Printf("[INFO] Reading network peering (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
	if err != nil {
		return diag.Errorf("unable to retrieve network peering (%s): %v", peeringID, err)
	}

	// Set the globally unique id of this peering in state.
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Store resource data.
	if err := setAwsPeeringResourceData(d, peering); err != nil {
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
	peering, err = clients.WaitForPeeringToBeActive(ctx, client, peering.ID, hvnID, loc, peeringCreateTimeout)
	if peering != nil {
		if err := setAwsPeeringResourceData(d, peering); err != nil {
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
