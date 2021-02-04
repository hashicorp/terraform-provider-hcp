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
var peeringDeleteTimeout = time.Minute * 35

func resourceAwsNetworkPeering() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS Network peering resource allows you to manage a Network peering between an HVN and a peer AWS VPC.",

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
			"peer_vpc_cidr_block": {
				Description:  "The CIDR range of the peer VPC in AWS.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			// Optional inputs
			"peering_id": {
				Description:      "The ID of the Network peering.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the Network peering is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the Network peering is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_peering_id": {
				Description: "The peering connection ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the Network peering was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the Network peering will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.",
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
			return diag.Errorf("unable to find the HVN (%s) for the Network peering", hvnID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnID, err)
	}
	log.Printf("[INFO] HVN (%s) found, proceeding with create", hvnID)

	// Check if peering already exists
	if peeringID != "" {
		_, err = clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
		if err != nil {
			if !clients.IsResponseCodeNotFound(err) {
				return diag.Errorf("unable to check for presence of an existing Network peering (%s): %v", peeringID, err)
			}

			log.Printf("[INFO] Network peering (%s) not found, proceeding with create", peeringID)
		} else {
			return diag.Errorf("a Network peering with peering_id=%s, hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the state. Please see the resource documentation for hcp_aws_network_peering for more information", peeringID, hvnID, loc.ProjectID)
		}
	}

	peerNetworkParams := network_service.NewCreatePeeringParams()
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
	log.Printf("[INFO] Creating Network peering between HVN (%s) and peer (%s)", hvnID, peerVpcID)
	peeringResponse, err := client.Network.CreatePeering(peerNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create Network peering between HVN (%s) and peer (%s): %v", hvnID, peerVpcID, err)
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

	// Wait for Network peering to be created
	if err := clients.WaitForOperation(ctx, client, "create Network peering", loc, peeringResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Network peering (%s) between HVN (%s) and peer (%s): %v", peering.ID, peering.Hvn.ID, peering.Target.AwsTarget.VpcID, err)
	}

	log.Printf("[INFO] Created Network peering (%s) between HVN (%s) and peer (%s)", peering.ID, peering.Hvn.ID, peering.Target.AwsTarget.VpcID)

	peering, err = clients.WaitForPeeringToBePendingAcceptance(ctx, client, peering.ID, hvnID, loc)
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

	log.Printf("[INFO] Reading Network peering (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve Network peering (%s): %v", peeringID, err)
	}

	// Network peering found, update resource data
	if err := setPeeringResourceData(d, peering); err != nil {
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
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvnID
	deletePeeringParams.LocationOrganizationID = loc.OrganizationID
	deletePeeringParams.LocationProjectID = loc.ProjectID
	log.Printf("[INFO] Deleting Network peering (%s)", peeringID)
	deletePeeringResponse, err := client.Network.DeletePeering(deletePeeringParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, so no action was taken", peeringID)
			return nil
		}

		return diag.Errorf("unable to delete Network peering (%s): %v", peeringID, err)
	}

	// Wait for peering to be deleted
	if err := clients.WaitForOperation(ctx, client, "delete Network peering", loc, deletePeeringResponse.Payload.Operation.ID); err != nil {
		// If the HVN has already been deleted
		// TODO: Peerings can be deleted automatically by the network monitor workflow ]
		// and when deleting the HVN is deleted can cause an already started error.
		// It would be nicer if we could return a better error message or tie the operations together.
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete Network peering (%s): %v", peeringID, err)
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

	return nil
}

// resourceAwsNetworkPeeringImport implements the logic necessary to import an
// un-tracked (by Terraform) Network peering resource into Terraform state.
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
