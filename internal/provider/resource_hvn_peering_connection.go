// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func resourceHvnPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "The HVN peering connection resource allows you to manage a peering connection between HVNs.",
		CreateContext: resourceHvnPeeringConnectionCreate,
		ReadContext:   resourceHvnPeeringConnectionRead,
		DeleteContext: resourceHvnPeeringConnectionDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
			Create:  &peeringCreateTimeout,
			Delete:  &peeringDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceHvnPeeringConnectionImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_1": {
				Description: "The unique URL of one of the HVNs being peered.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hvn_2": {
				Description: "The unique URL of one of the HVNs being peered.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Computed outputs
			"peering_id": {
				Description: "The ID of the peering connection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
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

func resourceHvnPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	orgID := client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      client.Config.ProjectID,
	}

	hvn1Link, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvn2Link, err := buildLinkFromURL(d.Get("hvn_2").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvn2, err := clients.GetHvnByID(ctx, client, hvn2Link.Location, hvn2Link.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	hvn2Link.Location.Region = &sharedmodels.HashicorpCloudLocationRegion{
		Provider: hvn2.Location.Region.Provider,
		Region:   hvn2.Location.Region.Region,
	}

	peerNetworkParams := network_service.NewCreatePeeringParams()
	peerNetworkParams.Context = ctx
	peerNetworkParams.PeeringHvnID = hvn1Link.ID
	peerNetworkParams.PeeringHvnLocationOrganizationID = hvn1Link.Location.OrganizationID
	peerNetworkParams.PeeringHvnLocationProjectID = hvn1Link.Location.ProjectID
	peerNetworkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreatePeeringRequest{
		Peering: &networkmodels.HashicorpCloudNetwork20200907Peering{
			Hvn: hvn1Link,
			Target: &networkmodels.HashicorpCloudNetwork20200907PeeringTarget{
				HvnTarget: &networkmodels.HashicorpCloudNetwork20200907NetworkTarget{
					Hvn: hvn2Link,
				},
			},
		},
	}
	log.Printf("[INFO] Creating peering connection between HVNs (%s), (%s)", hvn1Link.ID, hvn2Link.ID)
	peeringResponse, err := client.Network.CreatePeering(peerNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create peering connection between HVNs (%s) and (%s): %v", hvn1Link.ID, hvn1Link.ID, err)
	}

	peering := peeringResponse.Payload.Peering

	// Set the globally unique id of this peering in the state now since it has been created, and from this point forward should be deletable
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for peering connection to be created
	if err := clients.WaitForOperation(ctx, client, "create peering connection", loc, peeringResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create peering connection (%s) between HVNs (%s) and (%s): %v", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID, err)
	}
	log.Printf("[INFO] Created peering connection (%s) between HVNs (%s) and (%s)", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID)

	peering, err = clients.WaitForPeeringToBeAccepted(ctx, client, peering.ID, hvn1Link.ID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Peering connection (%s) is now in ACCEPTED state", peering.ID)

	if err := setHvnPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	orgID := client.Config.OrganizationID

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnLink1, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink1.ID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Peering connection (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	hvnLink2 := newLink(peering.Target.HvnTarget.Hvn.Location, HvnResourceType, peering.Target.HvnTarget.Hvn.ID)
	hvnURL2, err := linkURL(hvnLink2)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hvn_2", hvnURL2); err != nil {
		return diag.FromErr(err)
	}

	if err := setHvnPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	orgID := client.Config.OrganizationID

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnLink1, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	deletePeeringParams := network_service.NewDeletePeeringParams()
	deletePeeringParams.Context = ctx
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvnLink1.ID
	deletePeeringParams.LocationOrganizationID = loc.OrganizationID
	deletePeeringParams.LocationProjectID = loc.ProjectID

	log.Printf("[INFO] Deleting peering connection (%s)", peeringID)
	deletePeeringResponse, err := client.Network.DeletePeering(deletePeeringParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Peering connection (%s) not found, so no action was taken", peeringID)
			return nil
		}
		return diag.Errorf("unable to delete peering connection (%s): %v", peeringID, err)
	}

	// Wait for peering to be deleted
	if err := clients.WaitForOperation(ctx, client, "delete peering connection", loc, deletePeeringResponse.Payload.Operation.ID); err != nil {
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete peering connection (%s): %v", peeringID, err)
	}

	log.Printf("[INFO] Peering connection (%s) deleted, removing from state", peeringID)
	return nil
}

func resourceHvnPeeringConnectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)
	hvnID, peeringID, err := parsePeeringResourceID(d.Id())
	if err != nil {
		return nil, err
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	link := newLink(loc, PeeringResourceType, peeringID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	// Only hvn_1 is required to fetch the peering connection. hvn_2 will be populated during the refresh phase immediately after import.
	hvnLink := newLink(loc, HvnResourceType, hvnID)
	hvnURL, err := linkURL(hvnLink)
	if err != nil {
		return nil, err
	}

	d.SetId(url)
	if err := d.Set("hvn_1", hvnURL); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func setHvnPeeringResourceData(d *schema.ResourceData, peering *networkmodels.HashicorpCloudNetwork20200907Peering) error {
	if err := d.Set("peering_id", peering.ID); err != nil {
		return err
	}
	if err := d.Set("organization_id", peering.Hvn.Location.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", peering.Hvn.Location.ProjectID); err != nil {
		return err
	}
	if err := d.Set("created_at", peering.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("expires_at", peering.ExpiresAt.String()); err != nil {
		return err
	}
	if err := d.Set("state", peering.ExpiresAt.String()); err != nil {
		return err
	}

	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	return nil
}
