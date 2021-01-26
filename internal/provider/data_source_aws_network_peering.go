package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceAwsNetworkPeering() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS Network Peering data source provides information about an existing peering connection between an HVN and a peer AWS VPC.",
		ReadContext: dataSourceAwsNetworkPeeringRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &peeringDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"peering_id": {
				Description:      "The ID of the network peering.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the network peering is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the network peering is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_account_id": {
				Description: "The account ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vpc_id": {
				Description: "The ID of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vpc_region": {
				Description: "The region of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"peer_vpc_cidr_block": {
				Description: "The CIDR range of the peer VPC in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"provider_peering_id": {
				Description: "The peering connection ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the network peering was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the network peering will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceAwsNetworkPeeringRead is the func to implement reading of the
// AWS Network Peering connection between an HVN and a peer AWS VPC.
func dataSourceAwsNetworkPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}
	hvnID := d.Get("hvn_id").(string)
	peeringID := d.Get("peering_id").(string)

	log.Printf("[INFO] Reading network peering (%s)", peeringID)
	peering, err := clients.GetPeeringByID(ctx, client, peeringID, hvnID, loc)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Network peering (%s) not found, removing from state", peeringID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to retrieve network peering (%s): %v", peeringID, err)
	}

	// Network peering found, update resource data
	if err := setPeeringResourceData(d, peering); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
