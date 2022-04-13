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

func resourceAzurePeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description: "The Azure peering connection resource allows you to manage a peering connection between an HVN and a peer Azure VNet.",

		CreateContext: resourceAzurePeeringConnectionCreate,
		ReadContext:   resourceAzurePeeringConnectionRead,
		DeleteContext: resourceAzurePeeringConnectionDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
			Create:  &peeringCreateTimeout,
			Delete:  &peeringDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceAzurePeeringConnectionImport,
		},

		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_link": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peering_id": {
				Description:      "The ID of the peering connection.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"peer_vnet_name": {
				Description: "The name of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peer_subscription_id": {
				Description: "The subscription ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peer_vnet_region": {
				Description: "The region of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			"peer_tenant_id": {
				Description: "The tenant ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"peer_resource_group_name": {
				Description: "The resource group name of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				Description: "A unique URL identifying the peering connection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceAzurePeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	peeringID := d.Get("peering_id").(string)
	peerSubscriptionID := d.Get("peer_subscription_id").(string)
	peerVnetID := d.Get("peer_vnet_name").(string)
	peerVnetRegion := d.Get("peer_vnet_region").(string)
	peerTenantID := d.Get("peer_tenant_id").(string)
	peerResourceGroupName := d.Get("peer_resource_group_name").(string)

	orgID := client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      client.Config.ProjectID,
	}

	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check for an existing HVN
	_, err = clients.GetHvnByID(ctx, client, hvnLink.Location, hvnLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the peering connection", hvnLink.ID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnLink.ID, err)
	}
	log.Printf("[INFO] HVN (%s) found, proceeding with peering connection create", hvnLink.ID)

	// Check if peering already exists
	if peeringID != "" {
		_, err = clients.GetPeeringByID(ctx, client, peeringID, hvnLink.ID, loc)
		if err != nil {
			if !clients.IsResponseCodeNotFound(err) {
				return diag.Errorf("unable to check for presence of an existing peering connection (%s): %v", peeringID, err)
			}

			log.Printf("[INFO] peering connection (%s) not found, proceeding with peering connection create", peeringID)
		} else {
			return diag.Errorf("a peering connection with peering_id=%s, hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the state. Please see the resource documentation for hcp_Azure_network_peering for more information", peeringID, hvnLink.ID, loc.ProjectID)
		}
	}

	peerNetworkParams := network_service.NewCreatePeeringParams()
	peerNetworkParams.Context = ctx
	peerNetworkParams.PeeringHvnID = hvnLink.ID
	peerNetworkParams.PeeringHvnLocationOrganizationID = loc.OrganizationID
	peerNetworkParams.PeeringHvnLocationProjectID = loc.ProjectID
	peerNetworkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreatePeeringRequest{
		Peering: &networkmodels.HashicorpCloudNetwork20200907Peering{
			ID: peeringID,
			Hvn: &sharedmodels.HashicorpCloudLocationLink{
				ID:       hvnLink.ID,
				Location: loc,
			},
			Target: &networkmodels.HashicorpCloudNetwork20200907PeeringTarget{
				AzureTarget: &networkmodels.HashicorpCloudNetwork20200907AzurePeeringTarget{
					Region:            peerVnetRegion,
					ResourceGroupName: peerResourceGroupName,
					SubscriptionID:    peerSubscriptionID,
					TenantID:          peerTenantID,
					VnetName:          peerVnetID,
				},
			},
		},
	}
	log.Printf("[INFO] Creating peering connection between HVN (%s) and peer (%s)", hvnLink.ID, peerVnetID)
	peeringResponse, err := client.Network.CreatePeering(peerNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create peering connection between HVN (%s) and peer (%s): %v", hvnLink.ID, peerVnetID, err)
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

	peering, err = clients.WaitForPeeringToBePendingAcceptance(ctx, client, peering.ID, hvnLink.ID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] peering connection (%s) is now in PENDING_ACCEPTANCE state", peering.ID)

	if err := setAzurePeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAzurePeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location

	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, loc.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink.ID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] peering connection (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	// The peering connection failed to provision properly so we want to let the user know and
	// remove it from state
	if peering.State == networkmodels.HashicorpCloudNetwork20200907PeeringStateFAILED {
		log.Printf("[WARN] peering connection (%s) failed to provision, removing from state", peering.ID)
		d.SetId("")
		return nil
	}

	// peering connection found, update resource data
	if err := setAzurePeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAzurePeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), PeeringResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	peeringID := link.ID
	loc := link.Location
	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, loc.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	deletePeeringParams := network_service.NewDeletePeeringParams()
	deletePeeringParams.Context = ctx
	deletePeeringParams.ID = peeringID
	deletePeeringParams.HvnID = hvnLink.ID
	deletePeeringParams.LocationOrganizationID = loc.OrganizationID
	deletePeeringParams.LocationProjectID = loc.ProjectID
	log.Printf("[INFO] Deleting peering connection (%s)", peeringID)
	deletePeeringResponse, err := client.Network.DeletePeering(deletePeeringParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] peering connection (%s) not found, so no action was taken", peeringID)
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

	log.Printf("[INFO] peering connection (%s) deleted, removing from state", peeringID)

	return nil
}

func setAzurePeeringResourceData(d *schema.ResourceData, peering *networkmodels.HashicorpCloudNetwork20200907Peering) error {
	if err := d.Set("organization_id", peering.Hvn.Location.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", peering.Hvn.Location.ProjectID); err != nil {
		return err
	}
	if err := d.Set("peering_id", peering.ID); err != nil {
		return err
	}
	if err := d.Set("peer_subscription_id", peering.Target.AzureTarget.SubscriptionID); err != nil {
		return err
	}
	if err := d.Set("peer_vnet_name", peering.Target.AzureTarget.VnetName); err != nil {
		return err
	}
	if err := d.Set("peer_vnet_region", peering.Target.AzureTarget.Region); err != nil {
		return err
	}
	if err := d.Set("peer_resource_group_name", peering.Target.AzureTarget.ResourceGroupName); err != nil {
		return err
	}
	if err := d.Set("peer_tenant_id", peering.Target.AzureTarget.TenantID); err != nil {
		return err
	}
	if err := d.Set("azure_peering_id", peering.ProviderPeeringID); err != nil {
		return err
	}
	if err := d.Set("application_id", peering.Target.AzureTarget.ApplicationID); err != nil {
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

// resourceAzurePeeringConnectionImport implements the logic necessary to import an
// un-tracked (by Terraform) peering connection resource into Terraform state.
func resourceAzurePeeringConnectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	hvnLink := newLink(loc, HvnResourceType, hvnID)
	hvnUrl, err := linkURL(hvnLink)
	if err != nil {
		return nil, err
	}

	d.SetId(url)
	if err := d.Set("hvn_link", hvnUrl); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
