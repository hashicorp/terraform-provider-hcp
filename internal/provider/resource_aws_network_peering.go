package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var peeringDefaultTimeout = time.Minute * 1
var peeringCreateTimeout = time.Minute * 35
var peeringUpdateTimeout = time.Minute * 35
var peeringDeleteTimeout = time.Minute * 35

func resourceAwsNetworkPeering() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS network peering resource allows you to manage a network peering between an HVN and a peer AWS VPC.",

		CreateContext: resourceAwsNetworkPeeringCreate,
		ReadContext:   resourceAwsNetworkPeeringRead,
		UpdateContext: resourceAwsNetworkPeeringUpdate,
		DeleteContext: resourceAwsNetworkPeeringDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
			Create:  &peeringCreateTimeout,
			Update:  &peeringUpdateTimeout,
			Delete:  &peeringDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceAwsNetworkPeeringImport,
		},

		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"peer_account_id": {
				Description: "The account ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peer_vpc_id": {
				Description: "The ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peer_vpc_region": {
				Description: "The region of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			// Optional inputs
			"peer_vpc_cidr_block": {
				Description:  "The CIDR range of the peer VPC in AWS.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.IsCIDR,
			},
			"peering_id": {
				Description:      "The ID of the network peering.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateSlugID,
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
		},
	}
}

func resourceAwsNetworkPeeringCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	peeringID := d.Get("peering_id").(string)
	hvnID := d.Get("hvn_id").(string)
	peerAccountID := d.Get("peer_account_id").(string)
	peerVpcID := d.Get("peer_vpc_id").(string)
	peerVpcRegion := d.Get("peer_vpc_region").(string)
	peerVpcCidr := d.Get("peer_vpc_cidr_block").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	// Check for an existing HVN
	_, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the network peering", hvnID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnID, err)
	}
	log.Printf("[INFO] HVN (%s) found, proceeding with network peering create", hvnID)

	// Check if peering already exists
	if peeringID != "" {
		_, err = clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
		if err != nil {
			if !clients.IsResponseCodeNotFound(err) {
				return diag.Errorf("unable to check for presence of an existing network peering (%s): %v", peeringID, err)
			}

			log.Printf("[INFO] Network peering (%s) not found, proceeding with network peering create", peeringID)
		} else {
			return diag.Errorf("a network peering with peering_id=%s, hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the state. Please see the resource documentation for hcp_aws_network_peering for more information", peeringID, hvnID, loc.ProjectID)
		}
	}

	peerNetworkParams := network_service.NewCreatePeeringParams()
	peerNetworkParams.Context = ctx
	peerNetworkParams.PeeringHvnID = hvnID
	peerNetworkParams.PeeringHvnLocationOrganizationID = loc.OrganizationID
	peerNetworkParams.PeeringHvnLocationProjectID = loc.ProjectID
	peerNetworkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreatePeeringRequest{
		Peering: &networkmodels.HashicorpCloudNetwork20200907Peering{
			ID: peeringID,
			Hvn: &sharedmodels.HashicorpCloudLocationLink{
				ID:       hvnID,
				Location: loc,
			},
			Target: &networkmodels.HashicorpCloudNetwork20200907PeeringTarget{
				AwsTarget: &networkmodels.HashicorpCloudNetwork20200907AWSPeeringTarget{
					AccountID: peerAccountID,
					VpcID:     peerVpcID,
					Region:    peerVpcRegion,
					Cidr:      peerVpcCidr,
				},
			},
		},
	}
	log.Printf("[INFO] Creating network peering between HVN (%s) and peer (%s)", hvnID, peerVpcID)
	peeringResponse, err := client.Network.CreatePeering(peerNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create network peering between HVN (%s) and peer (%s): %v", hvnID, peerVpcID, err)
	}

	peering := peeringResponse.Payload.Peering

	// Set the globally unique id of this peering in the state now since it has
	// been created, and from this point forward should be deletable
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for network peering to be created
	if err := clients.WaitForOperation(ctx, client, "create network peering", loc, peeringResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create network peering (%s) between HVN (%s) and peer (%s): %v", peering.ID, peering.Hvn.ID, peering.Target.AwsTarget.VpcID, err)
	}

	log.Printf("[INFO] Created network peering (%s) between HVN (%s) and peer (%s)", peering.ID, peering.Hvn.ID, peering.Target.AwsTarget.VpcID)

	peering, err = clients.WaitForPeeringToBePendingAcceptance(ctx, client, peering.ID, hvnID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Network peering (%s) is now in PENDING_ACCEPTANCE state", peering.ID)

	if err := setPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsNetworkPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnID := d.Get("hvn_id").(string)

	log.Printf("[INFO] Reading network peering (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve network peering (%s): %v", peeringID, err)
	}

	// The network peering failed to provision properly so we want to let the user know and
	// remove it from state
	if peering.State == networkmodels.HashicorpCloudNetwork20200907PeeringStateFAILED {
		log.Printf("[WARN] Network peering (%s) failed to provision, removing from state", peering.ID)
		d.SetId("")
		return nil
	}

	// Network peering found, update resource data
	if err := setPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsNetworkPeeringUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnID := d.Get("hvn_id").(string)

	log.Printf("[INFO] Reading network peering for update(%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve network peering (%s): %v", peeringID, err)
	}

	// The network peering failed to provision properly so we want to let the user know and
	// remove it from state
	if peering.State == networkmodels.HashicorpCloudNetwork20200907PeeringStateFAILED {
		log.Printf("[WARN] Network peering (%s) failed to provision, removing from state", peering.ID)
		d.SetId("")
		return nil
	}

	oldCIDR, newCIDR := d.GetChange("peer_vpc_cidr_block")

	// If the field peer_vpc_cidr_block previously has a value, but is is not present for the update,
	// the HVN route for the peering will be removed.
	if newCIDR == "" && oldCIDR != "" {
		log.Printf("[WARN] Updating network peering (%s) route without a value for field peer_vpc_cid_block; this means the HVN route for this peering will be removed. To avoid deleting the peering's HVN route, please include peer_vpc_cidr_block = \"%v\" in the peering resource.", peering.ID, oldCIDR)

		route, err := clients.ListHVNRoutes(ctx, client, peering.Hvn.ID, oldCIDR.(string), "", "", loc)
		if err != nil {
			return diag.Errorf("unable to retrieve HVN route for Network peering (%s) in preparation for update", peeringID)
		}

		// ListHVNRoutes should only ever return 1 route for a specified HVN and destination CIDR.
		if len(route) != 1 {
			return diag.Errorf("unexpected number of HVN route retrieved for Network peering (%s): %v", peeringID, len(route))
		}

		resp, err := clients.DeleteHVNRouteByID(ctx, client, peering.Hvn.ID, route[0].ID, loc)
		log.Printf("operation ID for delete HVN route is %+v", resp.Operation.ID)
		if err := clients.WaitForOperation(ctx, client, "delete network peering HVN route", loc, resp.Operation.ID); err != nil {
			if strings.Contains(err.Error(), "execution already started") {
				return nil
			}
			return diag.Errorf("unable to delete network peering (%s) HVN route (%s): %v", peeringID, route[0].ID, err)
		}
	}

	// If the value for peer_vpc_cidr_block is being updated to a different non-empty value, return an error.
	if newCIDR != oldCIDR && newCIDR != "" {
		return diag.Errorf("Cannot update peer_vpc_cidr_block for Network peering (%s); peering route must be updated using HVN Route resource instead", peeringID)
	}

	// Set state for the peering
	return resourceAwsNetworkPeeringRead(ctx, d, meta)
}

func resourceAwsNetworkPeeringDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnID := d.Get("hvn_id").(string)

	deletePeeringParams := network_service.NewDeletePeeringParams()
	deletePeeringParams.Context = ctx
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvnID
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
		// If the HVN has already been deleted
		// TODO: Peerings can be deleted automatically by the network monitor workflow ]
		// and when deleting the HVN is deleted can cause an already started error.
		// It would be nicer if we could return a better error message or tie the operations together.
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete network peering (%s): %v", peeringID, err)
	}

	log.Printf("[INFO] Network peering (%s) deleted, removing from state", peeringID)

	return nil
}

func setPeeringResourceData(d *schema.ResourceData, peering *networkmodels.HashicorpCloudNetwork20200907Peering) error {
	if err := d.Set("peering_id", peering.ID); err != nil {
		return err
	}
	if err := d.Set("peer_account_id", peering.Target.AwsTarget.AccountID); err != nil {
		return err
	}
	if err := d.Set("peer_vpc_id", peering.Target.AwsTarget.VpcID); err != nil {
		return err
	}
	if err := d.Set("peer_vpc_region", peering.Target.AwsTarget.Region); err != nil {
		return err
	}
	if err := d.Set("peer_vpc_cidr_block", peering.Target.AwsTarget.Cidr); err != nil {
		return err
	}
	if err := d.Set("organization_id", peering.Hvn.Location.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", peering.Hvn.Location.ProjectID); err != nil {
		return err
	}
	if err := d.Set("provider_peering_id", peering.ProviderPeeringID); err != nil {
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

// resourceAwsNetworkPeeringImport implements the logic necessary to import an
// un-tracked (by Terraform) network peering resource into Terraform state.
func resourceAwsNetworkPeeringImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	idParts := strings.SplitN(d.Id(), ":", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{peering_id}", d.Id())
	}
	hvnID := idParts[0]
	peeringID := idParts[1]
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
