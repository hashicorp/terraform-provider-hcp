// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/stable/2020-09-07/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var hvnDefaultTimeout = time.Minute * 1
var hvnCreateTimeout = time.Minute * 10
var hvnDeleteTimeout = time.Minute * 10

var hvnResourceCloudProviders = []string{
	"aws",
	// Available to internal users only
	"azure",
}

func resourceHvn() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN resource allows you to manage a HashiCorp Virtual Network in HCP.",

		CreateContext: resourceHvnCreate,
		ReadContext:   resourceHvnRead,
		DeleteContext: resourceHvnDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnDefaultTimeout,
			Create:  &hvnCreateTimeout,
			Delete:  &hvnDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceHvnImport,
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
			"cloud_provider": {
				Description:      "The provider where the HVN is located. The provider 'aws' is generally available and 'azure' is in public beta.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSlice(hvnResourceCloudProviders, true),
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"region": {
				Description: "The region where the HVN is located.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			// Optional inputs
			"cidr_block": {
				Description:      "The CIDR range of the HVN. If this is not provided, the service will provide a default value.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateCIDRBlock,
				Computed:         true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the HVN is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the HVN is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_account_id": {
				Description: "The provider account ID where the HVN is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the HVN was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the HVN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the HVN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceHvnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: d.Get("cloud_provider").(string),
			Region:   d.Get("region").(string),
		},
	}

	// Check for an existing HVN
	_, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing HVN (%s): %v", hvnID, err)
		}

		log.Printf("[INFO] HVN (%s) not found, proceeding with create", hvnID)
	} else {
		return diag.Errorf("unable to create HVN (%s) - an HVN with this ID already exists; see resource documentation for hcp_hvn for instructions on how to add an already existing HVN to the state", hvnID)
	}

	createNetworkParams := network_service.NewCreateParams()
	createNetworkParams.Context = ctx
	createNetworkParams.Body = &networkmodels.HashicorpCloudNetwork20200907CreateRequest{
		Network: &networkmodels.HashicorpCloudNetwork20200907Network{
			ID:        hvnID,
			CidrBlock: cidrBlock,
			Location:  loc,
		},
	}
	createNetworkParams.NetworkLocationOrganizationID = loc.OrganizationID
	createNetworkParams.NetworkLocationProjectID = loc.ProjectID
	log.Printf("[INFO] Creating HVN (%s)", hvnID)
	createNetworkResponse, err := client.Network.Create(createNetworkParams, nil)
	if err != nil {
		return diag.Errorf("unable to create HVN (%s): %v", hvnID, err)
	}

	link := newLink(loc, HvnResourceType, hvnID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for HVN to be created
	if err := clients.WaitForOperation(ctx, client, "create HVN", loc, createNetworkResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create HVN (%s): %v", createNetworkResponse.Payload.Network.ID, err)
	}

	log.Printf("[INFO] Created HVN (%s)", createNetworkResponse.Payload.Network.ID)

	// Get the updated HVN
	hvn, err := clients.GetHvnByID(ctx, client, loc, createNetworkResponse.Payload.Network.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN (%s): %v", createNetworkResponse.Payload.Network.ID, err)
	}

	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvnID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading HVN (%s) [project_id=%s, organization_id=%s]", hvnID, loc.ProjectID, loc.OrganizationID)
	hvn, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN (%s) not found, removing from state", hvnID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve HVN (%s): %v", hvnID, err)
	}

	// The HVN has already been deleted, remove from state.
	if *hvn.State == networkmodels.HashicorpCloudNetwork20200907NetworkStateDELETED {
		log.Printf("[WARN] HVN (%s) failed to provision, removing from state", hvnID)
		d.SetId("")
		return nil
	}

	// HVN found, update resource data
	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHvnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), HvnResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	hvnID := link.ID
	loc := link.Location

	deleteParams := network_service.NewDeleteParams()
	deleteParams.Context = ctx
	deleteParams.ID = hvnID
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationProjectID = loc.ProjectID
	log.Printf("[INFO] Deleting HVN (%s)", hvnID)
	deleteResponse, err := client.Network.Delete(deleteParams, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] HVN (%s) not found, so no action was taken", hvnID)
			return nil
		}

		return diag.Errorf("unable to delete HVN (%s): %v", hvnID, err)
	}

	// Wait for delete hvn operation
	if err := clients.WaitForOperation(ctx, client, "delete HVN", loc, deleteResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to delete HVN (%s): %v", hvnID, err)
	}

	log.Printf("[INFO] HVN (%s) deleted, removing from state", hvnID)

	return nil
}

func setHvnResourceData(d *schema.ResourceData, hvn *networkmodels.HashicorpCloudNetwork20200907Network) error {
	if err := d.Set("hvn_id", hvn.ID); err != nil {
		return err
	}
	if err := d.Set("cidr_block", hvn.CidrBlock); err != nil {
		return err
	}
	if err := d.Set("organization_id", hvn.Location.OrganizationID); err != nil {
		return err
	}
	if err := d.Set("project_id", hvn.Location.ProjectID); err != nil {
		return err
	}
	if err := d.Set("cloud_provider", hvn.Location.Region.Provider); err != nil {
		return err
	}
	if err := d.Set("region", hvn.Location.Region.Region); err != nil {
		return err
	}
	if err := d.Set("created_at", hvn.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("state", hvn.State); err != nil {
		return err
	}

	var providerAccountID string
	switch d.Get("cloud_provider") {
	case "aws":
		providerAccountID = hvn.ProviderNetworkData.AwsNetworkData.AccountID
	case "azure":
		// No equivalent field exposed in Azure HVNs at this time
		providerAccountID = ""
	}
	if err := d.Set("provider_account_id", providerAccountID); err != nil {
		return err
	}

	link := newLink(hvn.Location, HvnResourceType, hvn.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}
	return nil
}

// resourceHvnImport implements the logic necessary to import an un-tracked
// (by Terraform) HVN resource into Terraform state.
func resourceHvnImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	hvnID := d.Id()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	link := newLink(loc, HvnResourceType, hvnID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}
