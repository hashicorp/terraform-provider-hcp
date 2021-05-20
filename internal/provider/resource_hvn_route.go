package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceHVNRouteImport,
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
				Description:      "The ID of the HVN route.",
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
				Description: "A unique URL identifying the target of the HVN route. Examples of the target: [`aws_network_peering`](aws_network_peering.md), [`aws_transit_gateway_attachment`](aws_transit_gateway_attachment.md)",
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
		},
	}
}

func resourceHvnRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	destination := d.Get("destination_cidr").(string)
	hvnRouteID := d.Get("hvn_route_id").(string)

	hvn := d.Get("hvn").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, loc.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	target := d.Get("target_link").(string)
	targetLink, err := parseLinkURL(target, "")
	if err != nil {
		return diag.Errorf("unable to parse target_link for HVN route (%s): %v", hvnRouteID, err)
	}
	targetLink.Location.OrganizationID = loc.OrganizationID

	// Check for an existing HVN.
	retrievedHvn, err := clients.GetHvnByID(ctx, client, loc, hvnLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the HVN route", hvnLink.ID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnLink.ID, err)
	}

	log.Printf("[INFO] HVN (%s) found, proceeding with HVN route create", hvnLink.ID)

	targetLink.Location.Region = retrievedHvn.Location.Region

	// Create HVN route
	hvnRouteResp, err := clients.CreateHVNRoute(ctx, client, hvnRouteID, hvnLink, destination, targetLink, loc)
	if err != nil {
		return diag.FromErr(err)
	}
	hvnRoute := hvnRouteResp.Route

	// Set the globally unique id of this HVN route in the state now since it has
	// been created, and from this point forward should be deletable.
	link := newLink(hvnRoute.Hvn.Location, HVNRouteResourceType, hvnRoute.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for HVN route to be created.
	if err := clients.WaitForOperation(ctx, client, "create HVN route", loc, hvnRouteResp.Operation.ID); err != nil {
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

	idLink, err := parseLinkURL(d.Id(), HVNRouteResourceType)
	if err != nil {
		return diag.FromErr(err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	route, err := clients.GetHVNRoute(ctx, client, hvnLink.ID, idLink.ID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN route (%s), removing from state", idLink.ID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve HVN route (%s): %v", idLink.ID, err)
	}

	// The HVN route failed to provision properly so we want to let the user know and remove it from state.
	if route.State == networkmodels.HashicorpCloudNetwork20200907HVNRouteStateFAILED {
		log.Printf("[WARN] HVN route (%s) failed to provision, removing from state", idLink.ID)
		d.SetId("")
		return nil
	}

	// HVN route found, update resource data.
	if err := setHVNRouteResourceData(d, route, loc); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceHvnRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), HVNRouteResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID := link.ID
	loc := link.Location

	hvn := d.Get("hvn").(string)
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, loc.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting HVN route (%s)", routeID)
	resp, err := clients.DeleteHVNRouteByID(ctx, client, hvnLink.ID, routeID, loc)

	if err := clients.WaitForOperation(ctx, client, "delete HVN route", loc, resp.Operation.ID); err != nil {
		if strings.Contains(err.Error(), "execution already started") {
			return nil
		}
		return diag.Errorf("unable to delete HVN route (%s): %v", routeID, err)
	}

	log.Printf("[INFO] HVN route (%s) deleted, removing from state", routeID)

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

	if err := d.Set("hvn_route_id", route.ID); err != nil {
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

// resourceHVNRouteImport implements the logic necessary to import an
// un-tracked (by Terraform) HVN route resource into Terraform state.
func resourceHVNRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	idParts := strings.SplitN(d.Id(), ":", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{hvn_route_id}", d.Id())
	}
	hvnID := idParts[0]
	routeID := idParts[1]
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	routeLink := newLink(loc, HVNRouteResourceType, routeID)
	routeUrl, err := linkURL(routeLink)
	if err != nil {
		return nil, err
	}
	d.SetId(routeUrl)

	hvnLink := newLink(loc, HvnResourceType, hvnID)
	hvnUrl, err := linkURL(hvnLink)
	if err != nil {
		return nil, err
	}

	if err := d.Set("hvn", hvnUrl); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
