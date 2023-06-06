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
		CustomizeDiff: resourceHvnPeeringConnectionCustomizeDiff,
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
			"project_id": {
				Description: "The ID of the HCP project where HVN peering connection is located. Always matches hvn_1's project ID. Setting this attribute is deprecated, but it will remain usable in read-only form.",
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
				Description: "The ID of the HCP organization where the peering connection is located. Always matches both HVNs' organization ID.",
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

func resourceHvnPeeringConnectionCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Force project_id to match the project_id from hvn_1 if it has been manually overridden in configuration
	// When then project_id attribute's "Optional" property is removed after the deprecation period
	// ends, CustomizeDiff can be removed.
	if d.HasChange("project_id") {
		hvn1Link, err := parseLinkURL(d.Get("hvn_1").(string), HvnResourceType)
		if err != nil {
			return err
		}
		if err := d.SetNew("project_id", hvn1Link.Location.ProjectID); err != nil {
			return err
		}
	}

	return nil
}

func resourceHvnPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	hvn1Link, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvn2Link, err := buildLinkFromURL(d.Get("hvn_2").(string), HvnResourceType, hvn1Link.Location.OrganizationID)
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
	if err := clients.WaitForOperation(ctx, client, "create peering connection", hvn1Link.Location, peeringResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create peering connection (%s) between HVNs (%s) and (%s): %v", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID, err)
	}
	log.Printf("[INFO] Created peering connection (%s) between HVNs (%s) and (%s)", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID)

	peering, err = clients.WaitForPeeringToBeAccepted(ctx, client, peering.ID, hvn1Link.ID, hvn1Link.Location, d.Timeout(schema.TimeoutCreate))
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

	hvn1Link, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringLink, err := buildLinkFromURL(d.Id(), PeeringResourceType, hvn1Link.Location.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := peeringLink.ID
	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvn1Link.ID, peeringLink.Location)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Peering connection (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

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

	return nil
}

func resourceHvnPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn1Link, err := buildLinkFromURL(d.Get("hvn_1").(string), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringLink, err := buildLinkFromURL(d.Id(), PeeringResourceType, hvn1Link.Location.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := peeringLink.ID

	deletePeeringParams := network_service.NewDeletePeeringParams()
	deletePeeringParams.Context = ctx
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvn1Link.ID
	deletePeeringParams.LocationOrganizationID = peeringLink.Location.OrganizationID
	deletePeeringParams.LocationProjectID = peeringLink.Location.ProjectID

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
	if err := clients.WaitForOperation(ctx, client, "delete peering connection", peeringLink.Location, deletePeeringResponse.Payload.Operation.ID); err != nil {
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete peering connection (%s): %v", peeringID, err)
	}

	log.Printf("[INFO] Peering connection (%s) deleted, removing from state", peeringID)
	return nil
}

func resourceHvnPeeringConnectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_hvn_peering_connection.test {project_id}:{hvn_id}:{peering_id}
	// use default project ID from provider:
	//   terraform import hcp_hvn_peering_connection.test {hvn_id}:{peering_id}

	client := meta.(*clients.Client)
	projectID, hvnID, peeringID, err := parsePeeringResourceID(d.Id(), client.Config.ProjectID)
	if err != nil {
		return nil, err
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: projectID,
	}

	link := newLink(loc, PeeringResourceType, peeringID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	// Only hvn_1 is required to fetch the peering connection. hvn_2 will be populated during the refresh phase immediately after import.
	hvn1Link := newLink(loc, HvnResourceType, hvnID)
	hvn1URL, err := linkURL(hvn1Link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)
	if err := d.Set("hvn_1", hvn1URL); err != nil {
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
