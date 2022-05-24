package provider

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func resourceAwsNetworkPeering() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS network peering resource allows you to manage a network peering between an HVN and a peer AWS VPC.",

		CreateContext: resourceAwsNetworkPeeringCreate,
		ReadContext:   resourceAwsNetworkPeeringRead,
		DeleteContext: resourceAwsNetworkPeeringDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
			Create:  &peeringCreateTimeout,
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
			"peering_id": {
				Description:      "The ID of the network peering.",
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

	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	peeringID := d.Get("peering_id").(string)
	hvnID := d.Get("hvn_id").(string)
	peerAccountID := d.Get("peer_account_id").(string)
	peerVpcID := d.Get("peer_vpc_id").(string)
	peerVpcRegion := d.Get("peer_vpc_region").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	// Check for an existing HVN
	_, err = clients.GetHvnByID(ctx, client, loc, hvnID)
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

	if err := setAwsPeeringResourceData(d, peering); err != nil {
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
	if err := setAwsPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
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

func setAwsPeeringResourceData(d *schema.ResourceData, peering *networkmodels.HashicorpCloudNetwork20200907Peering) error {
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
