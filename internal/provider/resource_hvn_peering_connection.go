package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func resourceHvnPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "The HVN peering connection resource allows you to manage a network peering connection between HVNs.",
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
			"peering_id": {
				Description:      "The ID of the network peering",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"hvn_1": {
				Description: "The unique URL of one of the HVNs being peer connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hvn_2": {
				Description: "The unique URL of one of the HVNs being peer connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the network peering is located. Always matches the HVNs' organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the network peering is located. Always matches the HVNs' project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the network peering was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the network peering will be considered expired if it hasn't into `ACCEPTED` or `ACTIVE` state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the network peering",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceHvnPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	peeringID := d.Get("peering_id").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	hvn1, err := getHvn(d.Get("hvn_1").(string), ctx, loc, client)
	if err != nil {
		return diag.FromErr(err)
	}

	hvn2, err := getHvn(d.Get("hvn_2").(string), ctx, loc, client)
	if err != nil {
		return diag.FromErr(err)
	}

	peerNetworkParams := network_service.NewCreatePeeringParams()
	peerNetworkParams.Context = ctx
	peerNetworkParams.PeeringHvnID = hvn1.ID
	peerNetworkParams.PeeringHvnLocationOrganizationID = hvn1.Location.OrganizationID
	peerNetworkParams.PeeringHvnLocationProjectID = hvn1.Location.ProjectID
	peerNetworkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreatePeeringRequest{
		Peering: &networkmodels.HashicorpCloudNetwork20200907Peering{
			ID:  peeringID,
			Hvn: newLink(hvn1.Location, HvnResourceType, hvn1.ID),
			Target: &networkmodels.HashicorpCloudNetwork20200907PeeringTarget{
				HvnTarget: &networkmodels.HashicorpCloudNetwork20200907NetworkTarget{
					Hvn: newLink(hvn2.Location, HvnResourceType, hvn2.ID),
				},
			},
		},
	}
	log.Printf("[INFO] Creating network peering between HVNs (%s), (%s)", hvn1.ID, hvn2.ID)
	peeringResponse, err := client.Network.CreatePeering(peerNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create network peering between HVNs (%s) and (%s): %v", hvn1.ID, hvn2.ID, err)
	}

	peering := peeringResponse.Payload.Peering

	// Set the globally unique id of this peering in the state now since it has been created, and from this point forward should be deletable
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for network peering to be created
	if err := clients.WaitForOperation(ctx, client, "create network peering", loc, peeringResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create network peering (%s) between HVNs (%s) and (%s): %v", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID, err)
	}
	log.Printf("[INFO] Created network peering (%s) between HVNs (%s) and (%s)", peering.ID, peering.Hvn.ID, peering.Target.HvnTarget.Hvn.ID)

	peering, err = clients.WaitForPeeringToBeAccepted(ctx, client, peering.ID, hvn1.ID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Network peering (%s) is now in ACCEPTED state", peering.ID)

	if err := setHvnPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnLink1, err := parseLinkURL(d.Get("hvn_1").(string), HvnResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading network peering (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink1.ID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve network peering (%s): %v", peeringID, err)
	}

	if peering.State == networkmodels.HashicorpCloudNetwork20200907PeeringStateFAILED {
		log.Printf("[WARN] Network peering (%s) failed to provision, removing from state", peering.ID)
		d.SetId("")
		return nil
	}

	if err := setHvnPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnLink1, err := parseLinkURL(d.Get("hvn_1").(string), HvnResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	deletePeeringParams := network_service.NewDeletePeeringParams()
	deletePeeringParams.Context = ctx
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvnLink1.ID
	deletePeeringParams.LocationOrganizationID = loc.OrganizationID
	deletePeeringParams.LocationProjectID = loc.ProjectID

	log.Printf("[INFO] Deleting network peering (%s)", peeringID)
	deletePeeringResponse, err := client.Network.DeletePeering(deletePeeringParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, so no action was taken", peeringID)
			return nil
		}
		return diag.Errorf("unable to delete network peering (%s): %v", peeringID, err)
	}

	// Wait for peering to be deleted
	if err := clients.WaitForOperation(ctx, client, "delete network peering", loc, deletePeeringResponse.Payload.Operation.ID); err != nil {
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete network peering (%s): %v", peeringID, err)
	}

	log.Printf("[INFO] Network peering (%s) deleted, removing from state", peeringID)
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

	d.SetId(url)
	if err := d.Set("hvn_id", hvnID); err != nil {
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

func getHvn(hvnSelfLink string, ctx context.Context, loc *sharedmodels.HashicorpCloudLocationLocation, client *clients.Client) (*networkmodels.HashicorpCloudNetwork20200907Network, error) {
	hvnLink, err := parseLinkURL(hvnSelfLink, HvnResourceType)
	if err != nil {
		return nil, err
	}
	hvn, err := clients.GetHvnByID(ctx, client, loc, hvnLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return nil, fmt.Errorf("unable to find the HVN (%s) for the network peering", hvnLink.ID)
		}
		return nil, fmt.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnLink.ID, err)
	}
	return hvn, nil
}
