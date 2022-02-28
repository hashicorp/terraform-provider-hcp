package provider

import (
	"context"
	"log"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceAzurePeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description: "The Azure peering connection data source provides information about a peering connection between an HVN and a peer Azure VNet.",
		ReadContext: dataSourceAzurePeeringConnectionRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringCreateTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"peering_id": {
				Description: "The ID of the peering connection.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hvn_link": {
				Description: "The `self_link` of the HashiCorp Virtual Network (HVN).",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional inputs
			"wait_for_active_state": {
				Description: "If `true`, Terraform will wait for the peering connection to reach an `ACTIVE` state before continuing. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
			"peer_vnet_name": {
				Description: "The name of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_subscription_id": {
				Description: "The subscription ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vnet_region": {
				Description: "The region of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_tenant_id": {
				Description: "The tenant ID of the peer VNet in Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_resource_group_name": {
				Description: "The resource group name of the peer VNet in Azure.",
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
				Description: "A unique URL identifying the peering connection",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAzurePeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	orgID := client.Config.OrganizationID

	peeringID := d.Get("peering_id").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnLink, err := buildLinkFromURL(d.Get("hvn_link").(string), HvnResourceType, orgID)
	if err != nil {
		return diag.FromErr(err)
	}
	waitForActive := d.Get("wait_for_active_state").(bool)

	log.Printf("[INFO] Reading peering connection (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnLink.ID, loc)
	if err != nil {
		return diag.Errorf("unable to retrieve peering connection (%s): %v", peeringID, err)
	}

	if waitForActive && peering.State != networkmodels.HashicorpCloudNetwork20200907PeeringStateACTIVE {
		peering, err = clients.WaitForPeeringToBeActive(ctx, client, peering.ID, hvnLink.ID, loc, peeringCreateTimeout)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Peering connection found, update resource data.
	if err := setAzurePeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	// Set the globally unique id of this peering in the state.
	link := newLink(peering.Hvn.Location, PeeringResourceType, peering.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	return nil
}
