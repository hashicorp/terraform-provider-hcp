// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceAwsTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS transit gateway attachment data source provides information about an existing transit gateway attachment.",
		ReadContext: dataSourceAwsTransitGatewayAttachmentRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &tgwDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"transit_gateway_attachment_id": {
				Description:      "The user-settable name of the transit gateway attachment in HCP.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"wait_for_active_state": {
				Description: "If `true`, Terraform will wait for the transit gateway attachment to reach an `ACTIVE` state before continuing. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the transit gateway attachment is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the transit gateway attachment is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"transit_gateway_id": {
				Description: "The ID of the user-owned transit gateway in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_transit_gateway_attachment_id": {
				Description: "The transit gateway attachment ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the transit gateway attachment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the transit gateway attachment was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the transit gateway attachment will be considered expired if it hasn't transitioned into `ACCEPTED` or `ACTIVE` state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the transit gateway attachment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAwsTransitGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	tgwAttID := d.Get("transit_gateway_attachment_id").(string)
	waitForActive := d.Get("wait_for_active_state").(bool)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	tgwAtt, err := clients.GetTGWAttachmentByID(ctx, client, tgwAttID, hvnID, loc)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set ID for the resource in state.
	link := newLink(tgwAtt.Location, TgwAttachmentResourceType, tgwAtt.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Update TF state for the resource before waiting.
	if err := setTransitGatewayAttachmentResourceData(d, tgwAtt); err != nil {
		return diag.FromErr(err)
	}

	// Skip waiting.
	if !waitForActive || tgwAtt.State == models.HashicorpCloudNetwork20200907TGWAttachmentStateACTIVE {
		return nil
	}

	// If it's not in a state where it could later become ACTIVE, we're going to bail.
	terminalState := true
	for _, state := range clients.WaitForTGWAttachmentToBeActiveStates {
		if state == string(tgwAtt.State) {
			terminalState = false
			break
		}
	}

	// If it's not in a state that we should wait on, issue a warning and bail.
	if terminalState {
		return []diag.Diagnostic{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Attachment is in an unexpected state, connections may fail: %q", string(tgwAtt.State)),
			Detail:   "Expected a CREATING, PENDING_ACCEPTANCE, ACCEPTED, or ACTIVE state",
		}}
	}

	// Store resource data again, updating Peering state.
	var result []diag.Diagnostic
	tgwAtt, err = clients.WaitForTGWAttachmentToBeActive(ctx, client, tgwAttID, hvnID, loc, d.Timeout(schema.TimeoutDefault))
	if tgwAtt != nil {
		if err := setTransitGatewayAttachmentResourceData(d, tgwAtt); err != nil {
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
