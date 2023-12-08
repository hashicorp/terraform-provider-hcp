// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var hvnRouteDefaultTimeout = time.Minute * 1
var hvnRouteCreateTimeout = time.Minute * 35
var hvnRouteDeleteTimeout = time.Minute * 25

func resourceHvnRoute() *schema.Resource {
	return &schema.Resource{
		Description:   "The HVN route resource allows you to manage an HVN route.",
		CreateContext: resourceHvnRouteCreate,
		ReadContext:   resourceHvnRouteRead,
		DeleteContext: resourceHvnRouteDelete,
		CustomizeDiff: resourceHvnRouteCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnRouteDefaultTimeout,
			Create:  &hvnRouteCreateTimeout,
			Delete:  &hvnRouteDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceHVNRouteImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_link": {
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
				Description:      "The destination CIDR of the HVN route.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateCIDRBlockHVNRoute,
			},
			"target_link": {
				Description: "A unique URL identifying the target of the HVN route. Examples of the target: [`aws_network_peering`](aws_network_peering.md), [`aws_transit_gateway_attachment`](aws_transit_gateway_attachment.md)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Optional inputs
			"azure_config": {
				Description: "The azure configuration for routing.",
				Type:        schema.TypeList,
				ForceNew:    true,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"next_hop_type": {
							Description: "The type of Azure hop the packet should be sent to.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							ValidateFunc: validation.StringInSlice([]string{
								"VirtualAppliance",
								"VirtualNetworkGateway",
							}, true),
						},
						"next_hop_ip_address": {
							Description: "Contains the IP address packets should be forwarded to. Next hop values are only allowed in routes where the next hop type is VirtualAppliance.",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
			// Computed outputs
			"project_id": {
				Description: "The ID of the HCP project where the HVN route is located. Always matches the project ID in `hvn_link`. Setting this attribute is deprecated, but it will remain usable in read-only form.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Deprecated: `
Setting the 'project_id' attribute is deprecated, but it will remain usable in read-only form.
Previously, the value for this attribute was required to match the project ID contained in 'hvn_link'. Now, the value will be calculated automatically.
Remove this attribute from the configuration for any affected resources.
`,
			},
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

func resourceHvnRouteCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Force project_id to match the project_id from hvn_link if it has been manually overridden in configuration
	// When the project_id attribute's "Optional" property is removed after the deprecation period
	// ends, CustomizeDiff can be removed.
	if d.HasChange("project_id") {
		hvnLink, err := parseLinkURL(d.Get("hvn_link").(string), HvnResourceType)
		if err != nil {
			return err
		}
		if err := d.SetNew("project_id", hvnLink.Location.ProjectID); err != nil {
			return err
		}
	}

	return nil
}

func resourceHvnRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	destination := d.Get("destination_cidr").(string)
	hvnRouteID := d.Get("hvn_route_id").(string)

	hvn := d.Get("hvn_link").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	target := d.Get("target_link").(string)
	targetLink, err := parseLinkURL(target, "")
	if err != nil {
		return diag.Errorf("unable to parse target_link for HVN route (%s): %v", hvnRouteID, err)
	}
	targetLink.Location.OrganizationID = hvnLink.Location.OrganizationID

	// Get azure config
	azureConfig, diagErr := getAzureConfig("azure_config", d)
	if diagErr != nil {
		return diagErr
	}

	// Check for an existing HVN.
	retrievedHvn, err := clients.GetHvnByID(ctx, client, hvnLink.Location, hvnLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to find the HVN (%s) for the HVN route", hvnLink.ID)
		}

		return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnLink.ID, err)
	}

	log.Printf("[INFO] HVN (%s) found, proceeding with HVN route create", hvnLink.ID)

	targetLink.Location.Region = retrievedHvn.Location.Region

	// Create HVN route
	hvnRouteResp, err := clients.CreateHVNRoute(ctx, client, hvnRouteID, hvnLink, destination, targetLink, hvnLink.Location, azureConfig)
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
	if err := clients.WaitForOperation(ctx, client, "create HVN route", hvnLink.Location, hvnRouteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to create HVN route (%s): %v", hvnRouteID, err)
	}

	log.Printf("[INFO] Created HVN route (%s)", hvnRouteID)

	hvnRoute, err = clients.WaitForHVNRouteToBeActive(ctx, client, hvnLink.ID, hvnRouteID, hvnLink.Location, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setHVNRouteResourceData(d, hvnRoute, hvnLink.Location); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceHvnRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn := d.Get("hvn_link").(string)
	var hvnLink *sharedmodels.HashicorpCloudLocationLink

	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	routeLink, err := buildLinkFromURL(d.Id(), HVNRouteResourceType, hvnLink.Location.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading HVN route (%s)", routeLink.ID)
	route, err := clients.GetHVNRoute(ctx, client, hvnLink.ID, routeLink.ID, hvnLink.Location)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN route (%s) not found, removing from state", routeLink.ID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve HVN route (%s): %v", routeLink.ID, err)
	}

	// HVN route found, update resource data.
	if err := setHVNRouteResourceData(d, route, hvnLink.Location); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceHvnRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvn := d.Get("hvn_link").(string)
	hvnLink, err := buildLinkFromURL(hvn, HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	routeLink, err := buildLinkFromURL(d.Id(), HVNRouteResourceType, hvnLink.Location.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID := routeLink.ID

	log.Printf("[INFO] Deleting HVN route (%s)", routeID)
	resp, err := clients.DeleteHVNRouteByID(ctx, client, hvnLink.ID, routeID, hvnLink.Location)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN route (%s) not found, so no action was taken", routeID)
			return nil
		}

		return diag.Errorf("unable to delete HVN route (%s): %v", routeID, err)
	}

	if err := clients.WaitForOperation(ctx, client, "delete HVN route", hvnLink.Location, resp.Operation.ID); err != nil {
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

	if err := d.Set("project_id", loc.ProjectID); err != nil {
		return err
	}

	if err := d.Set("hvn_route_id", route.ID); err != nil {
		return err
	}

	if err := d.Set("target_link", targetLink); err != nil {
		return err
	}

	if err := d.Set("azure_config", flattenAzureConfig(route.AzureRoute, d)); err != nil {
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
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_hvn_route.test {project_id}:{hvn_id}:{hvn_route_id}
	// use default project ID from provider:
	//   terraform import hcp_hvn_route.test {hvn_id}:{hvn_route_id}

	client := meta.(*clients.Client)
	projectID := ""
	hvnID := ""
	routeID := ""
	var err error
	// Updates the source channel to include data about the module used.
	client, err = client.UpdateSourceChannel(d)
	if err != nil {
		log.Printf("[DEBUG] Failed to update analytics with module name (%s)", err)
	}

	idParts := strings.SplitN(d.Id(), ":", 3)
	if len(idParts) == 3 { // {project_id}:{hvn_id}:{hvn_route_id}
		if idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {project_id}:{hvn_id}:{hvn_route_id}", d.Id())
		}
		projectID = idParts[0]
		hvnID = idParts[1]
		routeID = idParts[2]
	} else if len(idParts) == 2 { // {hvn_id}:{hvn_route_id}
		if idParts[0] == "" || idParts[1] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{hvn_route_id}", d.Id())
		}
		projectID, err = GetProjectID(projectID, client.Config.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve project ID: %v", err)
		}
		hvnID = idParts[0]
		routeID = idParts[1]
	} else {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected {hvn_id}:{hvn_route_id} or {project_id}:{hvn_id}:{hvn_route_id}", d.Id())
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: projectID,
	}

	routeLink := newLink(loc, HVNRouteResourceType, routeID)
	routeURL, err := linkURL(routeLink)
	if err != nil {
		return nil, err
	}
	d.SetId(routeURL)

	hvnLink := newLink(loc, HvnResourceType, hvnID)
	hvnURL, err := linkURL(hvnLink)
	if err != nil {
		return nil, err
	}

	if err := d.Set("hvn_link", hvnURL); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func getAzureConfig(propertyName string, d *schema.ResourceData) (*networkmodels.HashicorpCloudNetwork20200907AzureRoute, diag.Diagnostics) {
	emptyConfig := networkmodels.HashicorpCloudNetwork20200907AzureRoute{}

	// If we don't find the property we return the empty object to be updated and delete the configuration.
	configParam, ok := d.GetOk(propertyName)
	if !ok {
		return &emptyConfig, nil
	}
	configIfaceArr, ok := configParam.([]interface{})
	if !ok || len(configIfaceArr) == 0 {
		return &emptyConfig, nil
	}
	config, ok := configIfaceArr[0].(map[string]interface{})
	if !ok {
		return &emptyConfig, nil
	}

	return getValidAzureRouteConfig(config)
}

func getValidAzureRouteConfig(config map[string]interface{}) (*networkmodels.HashicorpCloudNetwork20200907AzureRoute, diag.Diagnostics) {
	nextHopType, _ := config["next_hop_type"].(string)
	nextHopIPAddress, _ := config["next_hop_ip_address"].(string)

	// Verify if next_hop_ip_address is set, next_hop_type is VirtualAppliance
	if nextHopIPAddress != "" && strings.ToUpper(nextHopType) != "VIRTUALAPPLIANCE" {
		return nil, diag.Errorf("azure configuration is invalid: Next hop values are only allowed in routes where next hp type is VirtualAppliance")
	}

	return &networkmodels.HashicorpCloudNetwork20200907AzureRoute{
		NextHopIPAddress: nextHopIPAddress,
		NextHopType:      (*networkmodels.HashicorpCloudNetwork20200907AzureHopType)(&nextHopType),
	}, nil
}

func flattenAzureConfig(config *networkmodels.HashicorpCloudNetwork20200907AzureRoute, d *schema.ResourceData) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"next_hop_type":       config.NextHopType,
			"next_hop_ip_address": config.NextHopIPAddress,
		},
	}
}
