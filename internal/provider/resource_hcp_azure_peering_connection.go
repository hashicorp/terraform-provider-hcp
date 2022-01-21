package provider

import (
	"context"
	"strings"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"hvn": {
				Description: "The unique URL of the HashiCorp Virtual Network (HVN).",
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
			"peer_vnet_id": {
				Description: "The ID of the peer VNet in Azure.",
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
					return strings.EqualFold(old, new)
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
				Description: "The application ID of the HCP VNet backing the HVN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_peering_id": {
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
	return nil
}

func resourceAzurePeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceAzurePeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func setAzurePeeringResourceData(d *schema.ResourceData, peering *networkmodels.HashicorpCloudNetwork20200907Peering) error {
	return nil
}

// resourceAzurePeeringConnectionImport implements the logic necessary to import an
// un-tracked (by Terraform) peering connection resource into Terraform state.
func resourceAzurePeeringConnectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
