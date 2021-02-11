package provider

import (
	"context"

	networkmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-network/preview/2020-09-07/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceAwsTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "The AWS Transit Gateway Attachment data source provides information about an existing transit gateway attachment.",
		ReadContext: dataSourceAwsTransitGatewayAttachmentRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &tgwDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"transit_gateway_attachment_id": {
				Description:      "The user-settable name of the transit gateway attachment in HCP.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"wait_for_active_state": {
				Description: "If `true`, Terraform will wait for the transit gateway attachment to reach an `ACTIVE` state before continuing. Default `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the transit gateway attachment is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the transit gateway attachment is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"transit_gateway_id": {
				Description: "The ID of the user-owned transit gateway in AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"destination_cidrs": {
				Description: "The list of associated CIDR ranges. Traffic from these CIDRs will be allowed for all resources in the HVN. Traffic to these CIDRs will be routed into this transit gateway attachment.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"provider_transit_gateway_attachment_id": {
				Description: "The transit gateway attachment ID used by AWS.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the transit gateway attachment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the transit gateway attachment was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the transit gateway attachment will be considered expired if it hasn't transitioned into `ACCEPTED` or `ACTIVE` state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAwsTransitGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	tgwAttID := d.Get("transit_gateway_attachment_id").(string)
	waitForActive := d.Get("wait_for_active_state").(bool)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	tgwAtt, err := clients.GetTGWAttachmentByID(ctx, client, tgwAttID, hvnID, loc)
	if err != nil {
		return diag.FromErr(err)
	}

	if waitForActive && tgwAtt.State != networkmodels.HashicorpCloudNetwork20200907TGWAttachmentStateACTIVE {
		tgwAtt, err = clients.WaitForTGWAttachmentToBeActive(ctx, client, tgwAttID, hvnID, loc, d.Timeout(schema.TimeoutDefault))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	link := newLink(tgwAtt.Location, TgwAttachmentResourceType, tgwAtt.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	if err := setTransitGatewayAttachmentResourceData(d, tgwAtt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
