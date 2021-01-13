package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/client/network_service"
	networkmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var hvnDefaultTimeout = time.Minute * 1
var hvnCreateTimeout = time.Minute * 10
var hvnDeleteTimeout = time.Minute * 10

func resourceHcpHvn() *schema.Resource {
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
			"hvn_id": {
				Description: "The ID of the HashiCorp Virtual Network.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cidr_block": {
				Description:  "The CIDR range of the HVN.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"organization_id": {
				Description: "The ID of the HCP organization where the HVN is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the HVN is located.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"cloud_provider": {
				Description: "The provider where the HVN is located. Only 'aws' is available at this time.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws",
				}, true),
			},
			"region": {
				Description: "The region where the HVN is located.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceHcpHvnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	cidrBlock := d.Get("cidr_block").(string)

	loc, err := buildHvnResourceLocation(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
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
		return diag.Errorf("unable to create HVN [id=%s]: %+v", hvnID, err)
	}

	d.SetId(createNetworkResponse.Payload.Network.ID)

	// Get the updated HVN
	hvn, err := clients.GetHvnByID(ctx, client, loc, createNetworkResponse.Payload.Network.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve HVN [id=%s]: %+v", createNetworkResponse.Payload.Network.ID, err)
	}
	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	// Wait for HVN to be created
	if err := clients.WaitForOperation(ctx, client, "create HVN", loc, createNetworkResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create HVN [id=%s]: %+v", hvn.ID, err)
	}

	log.Printf("[INFO] Created HVN (%s)", hvn.ID)

	return nil
}

func resourceHcpHvnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Id()

	loc, err := buildHvnResourceLocation(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading HVN (%s) [project_id=%s, organization_id=%s]", hvnID, loc.ProjectID, loc.OrganizationID)

	hvn, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		// Is the hvn not found
		var apiErr *runtime.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			log.Printf("[WARN] HVN (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve HVN [id=%s]: %+v", hvnID, err)
	}

	// HVN found, update resource data
	if err := setHvnResourceData(d, hvn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHcpHvnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Id()

	loc, err := buildHvnResourceLocation(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	deleteParams := network_service.NewDeleteParams()
	deleteParams.Context = ctx
	deleteParams.ID = hvnID
	deleteParams.LocationOrganizationID = loc.OrganizationID
	deleteParams.LocationProjectID = loc.ProjectID
	deleteParams.LocationRegionProvider = &loc.Region.Provider
	deleteParams.LocationRegionRegion = &loc.Region.Region
	log.Printf("[INFO] Deleting HVN: [id=%s]", hvnID)
	deleteResponse, err := client.Network.Delete(deleteParams, nil)
	if err != nil {
		return diag.Errorf("unable to delete HVN [id=%s]: %+v", hvnID, err)
	}

	// Wait for delete hvn operation
	if err := clients.WaitForOperation(ctx, client, "delete HVN", loc, deleteResponse.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to delete HVN [id=%s]: %+v", hvnID, err)
	}

	log.Printf("[INFO] HVN (%s) deleted, removing from state", d.Id())
	d.SetId("")

	return nil
}

func setHvnResourceData(d *schema.ResourceData, hvn *networkmodels.HashicorpCloudNetwork20200907Network) error {
	d.SetId(hvn.ID)
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
	return nil
}

func buildHvnResourceLocation(ctx context.Context, d *schema.ResourceData, client *clients.Client) (*sharedmodels.HashicorpCloudLocationLocation, error) {
	provider := d.Get("cloud_provider").(string)
	region := d.Get("region").(string)

	projectID := client.Config.ProjectID
	organizationID := client.Config.OrganizationID
	projectIDVal, ok := d.GetOk("project_id")
	if ok {
		projectID = projectIDVal.(string)
		// Try to get organization_id from state, since project_id might have come from state
		organizationID = d.Get("organization_id").(string)
	}

	if projectID == "" {
		return nil, fmt.Errorf("missing project_id: a project_id must be specified on the HVN resource or the provider")
	}

	if organizationID == "" {
		var err error
		organizationID, err = clients.GetParentOrganizationIDByProjectID(ctx, client, projectID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve organization ID for project [project_id=%s]: %+v", projectID, err)
		}
	}

	return &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: provider,
			Region:   region,
		},
	}, nil
}
