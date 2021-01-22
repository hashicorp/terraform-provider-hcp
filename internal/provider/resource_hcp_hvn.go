package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/helper"
)

var hvnDefaultTimeout = time.Minute * 1
var hvnCreateTimeout = time.Minute * 10
var hvnDeleteTimeout = time.Minute * 10

var hvnResourceCloudProviders = []string{
	"aws",
}

func resourceHvn() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN resource allows you to manage a HashiCorp Virtual Network in HCP.",

		CreateContext: resourceHcpHvnCreate,
		ReadContext:   resourceHcpHvnRead,
		DeleteContext: resourceHcpHvnDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnDefaultTimeout,
			Create:  &hvnCreateTimeout,
			Delete:  &hvnDeleteTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description: "The ID of the HashiCorp Virtual Network.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cloud_provider": {
				Description:  "The provider where the HVN is located. Only 'aws' is available at this time.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(hvnResourceCloudProviders, true),
			},
			"region": {
				Description: "The region where the HVN is located.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Optional inputs
			"cidr_block": {
				Description:  "The CIDR range of the HVN. If this is not provided, the service will provide a default value.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
				Computed:     true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the HVN is located.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the HVN is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the HVN was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceHcpHvnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	loc, err := helper.BuildResourceLocationWithRegion(ctx, d, client, "HVN")
	if err != nil {
		return diag.FromErr(err)
	}

	// Check for an existing HVN
	_, err = clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing HVN (%s): %+v", hvnID, err)
		}

		log.Printf("[INFO] HVN (%s) not found, proceeding with create", hvnID)
	} else {
		return diag.Errorf("an HVN with hvn_id=%s and project_id=%s already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for hcp_hvn for more information", hvnID, loc.ProjectID)
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
		return diag.Errorf("unable to create HVN (%s): %+v", hvnID, err)
	}

	// Wait for HVN to be created
	if err := clients.WaitForOperation(ctx, client, "create HVN", loc, createNetworkResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create HVN (%s): %+v", createNetworkResponse.Payload.Network.ID, err)
	}

	log.Printf("[INFO] Created HVN (%s)", createNetworkResponse.Payload.Network.ID)

	// Get the updated HVN
	hvn, err := clients.GetHvnByID(ctx, client, loc, createNetworkResponse.Payload.Network.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN (%s): %+v", createNetworkResponse.Payload.Network.ID, err)
	}

	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHcpHvnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := parseLinkURL(d.Id())
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

		return diag.Errorf("unable to retrieve HVN (%s): %+v", hvnID, err)
	}

	// HVN found, update resource data
	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHcpHvnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := parseLinkURL(d.Id())
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

		return diag.Errorf("unable to delete HVN (%s): %+v", hvnID, err)
	}

	// Wait for delete hvn operation
	if err := clients.WaitForOperation(ctx, client, "delete HVN", loc, deleteResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to delete HVN (%s): %+v", hvnID, err)
	}

	log.Printf("[INFO] HVN (%s) deleted, removing from state", hvnID)
	d.SetId("")

	return nil
}

func setHvnResourceData(d *schema.ResourceData, hvn *networkmodels.HashicorpCloudNetwork20200907Network) error {
	link := newLink(hvn.Location, "hvn", hvn.ID)
	url, err := linkURL(link)
	if err != nil {
		return err
	}
	d.SetId(url)

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
	return nil
}
