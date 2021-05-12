package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var hvnRouteDefaultTimeout = time.Minute * 1
var hvnRouteCreateTimeout = time.Minute * 35

func resourceHvnRoute() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN route resource allows you to manage an HVN route.",

		CreateContext: resourceHvnRouteCreate,
		ReadContext:   resourceHvnRouteRead,
		DeleteContext: resourceHvnRouteDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnRouteDefaultTimeout,
			Create:  &hvnRouteCreateTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hvn_route_id": {
				Description:      "The ID of the network peering.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"destination_cidr": {
				Description:  "The destination CIDR of the HVN route",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"target_link": {
				Description: "A unique URL identifying the target of the HVN route.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Computed outputs
			"self_link": {
				Description: "A unique URL identifying the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the HVN route.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the HVN route was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// TODO add updated_at after public API returns updated_at for HVN route
		},
	}
}

func resourceHvnRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	hvn := d.Get("hvn").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, loc.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	target := d.Get("target_link").(string)
	var targetLink *sharedmodels.HashicorpCloudLocationLink
	// TODO this needs to handle more than just peering type
	targetLink, err = buildLinkFromURL(target, PeeringResourceType, loc.OrganizationID)

	destination := d.Get("destination_cidr").(string)
	hvnRouteID := d.Get("hvn_route_id").(string)
	// Check for an existing HVN.
	retrievedHvn, err := clients.GetHvnByID(ctx, client, loc, hvnLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the HVN route", hvnLink.ID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnLink.ID, err)
	}

	if retrievedHvn.State != networkmodels.HashicorpCloudNetwork20200907NetworkStateSTABLE {
		return diag.Errorf("unable to create HVN route for HVN (%s), HVN state is not STABLE", retrievedHvn.ID)
	}

	log.Printf("[INFO] HVN (%s) found, proceeding with HVN route create", hvnLink.ID)

	// Check if HVN route already exists.
	_, err = clients.ListHVNRoutes(ctx, client, hvnLink.ID, destination, "", "", loc)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing route for HVN (%s) with the destination CIDR of %s: %v", hvnLink.ID, destination, err)
		}

		return diag.Errorf("an HVN route with destination=%s, hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the state. Please see the resource documentation for hcp_hvn_route for more information", destination, hvnLink.ID, loc.ProjectID)
	} else {
		log.Printf("[INFO] HVN route with destination CIDR of %s for HVN (%s) not found, proceeding with HVN route create", destination, hvnLink.ID)
	}

	// TODO Check if the target exists and is in operational state.

	targetLink.Location.Region = retrievedHvn.Location.Region

	log.Printf("targetLink is %+v, hvnlink is %+v", targetLink.Location, hvnLink.Location)
	hvnRouteParams := network_service.NewCreateHVNRouteParams()
	hvnRouteParams.Context = ctx
	hvnRouteParams.HvnLocationOrganizationID = loc.OrganizationID
	hvnRouteParams.HvnLocationProjectID = loc.ProjectID
	hvnRouteParams.HvnID = hvnLink.ID
	hvnRouteParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateHVNRouteRequest{
		Destination: destination,
		Hvn:         hvnLink,
		ID:          hvnRouteID,
		Target: &networkmodels.HashicorpCloudNetwork20200907HVNRouteTarget{
			HvnConnection: targetLink,
		},
	}
	log.Printf("[INFO] Creating HVN route for HVN (%s) with destination CIDR %s", hvnLink.ID, destination)
	hvnRouteResp, err := client.Network.CreateHVNRoute(hvnRouteParams, nil)
	if err != nil {
		return diag.Errorf("unable to create HVN route for HVN (%s) with destination CIDR %s: %v", hvnLink.ID, destination, err)
	}

	hvnRoute := hvnRouteResp.Payload.Route

	// Set the globally unique id of this HVN route in the state now since it has
	// been created, and from this point forward should be deletable.
	link := newLink(hvnRoute.Hvn.Location, HVNRouteResourceType, hvnRoute.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for HVN route to be created.
	if err := clients.WaitForOperation(ctx, client, "create HVN route", loc, hvnRouteResp.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create HVN route for HVN (%s) with destination CIDR %s:%v", hvnLink.ID, destination, err)
	}

	log.Printf("[INFO] Created HVN route for HVN (%s) with destination CIDR %s", hvnRoute.ID, hvnRoute.Hvn.ID)

	hvnRoute, err = clients.WaitForHVNRouteToBeActive(ctx, client, destination, hvnLink.ID, loc, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setHVNRouteResourceData(d, hvnRoute, loc); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceHvnRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn := d.Get("hvn").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink

	hvnLink, err := parseLinkURL(hvn, HvnResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	destination := d.Get("destination").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	route, err := clients.ListHVNRoutes(ctx, client, hvnLink.ID, destination, "", "", loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN route for HVN (%s) with destination CIDR %s not found, removing from state", hvnLink.ID, destination)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve HVN route for HVN (%s) with destination CIDR %s: %v", hvnLink.ID, destination, err)
	}

	if len(route) != 1 {
		return diag.Errorf("Unexpected number of HVN route returned when waiting for route with destination CIDR of %s for HVN (%s) to be Active", destination, hvnLink.ID)
	}

	// The HVN route failed to provision properly so we want to let the user know and
	// remove it from state.
	if route[0].State == networkmodels.HashicorpCloudNetwork20200907HVNRouteStateFAILED {
		log.Printf("[WARN] HVN route for HVN (%s) with destination CIDR %s failed to provision, removing from state", hvnLink.ID, destination)
		d.SetId("")
		return nil
	}

	// HVN route found, update resource data.
	if err := setHVNRouteResourceData(d, route[0], loc); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceHvnRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func setHVNRouteResourceData(d *schema.ResourceData, route *networkmodels.HashicorpCloudNetwork20200907HVNRoute,
	loc *sharedmodels.HashicorpCloudLocationLocation) error {

	// Set self_link for the HVN route.
	link := newLink(loc, HVNRouteResourceType, route.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}

	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	// Set self_link identifying the target of the HVN route.
	hvnLink := newLink(loc, route.Target.HvnConnection.Type, route.Target.HvnConnection.ID)
	targetLink, err := linkURL(hvnLink)
	if err != nil {
		return err
	}

	if err := d.Set("target_link", targetLink); err != nil {
		return err
	}

	if err := d.Set("destination_cidr", route.Destination); err != nil {
		return err
	}

	if err := d.Set("state", route.State); err != nil {
		return err
	}

	if err := d.Set("created_at", route.CreatedAt.String()); err != nil {
		return err
	}

	return nil
}
