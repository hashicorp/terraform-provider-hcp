package provider

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceTGWAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "??????",
		ReadContext: dataSourceTGWAttachmentRead,
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
			"tgw_attachment_id": {
				Description: "The ID of the Transit Gateway (TGW) attachment.",
				Type:        schema.TypeString,
				Required:    true,
				// ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"state": {
				Description: "?????????",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the TGW attachment is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the TGW attachment is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tgw_id": {
				Description: "?????????",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"destination_cidrs": {
				Description: "?????????",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"provider_tgw_attachment_id": {
				Description: "?????????",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the TGW attachment was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The time after which the TGW attachment will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTGWAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)
	tgwAttID := d.Get("tgw_attachment_id").(string)
	tgwAttState := d.Get("state").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	tgwAtt, err := clients.GetTGWAttachmentByID(ctx, client, tgwAttID, hvnID, loc)
	if err == nil && tgwAttState != "" {
		tgwAtt, err = clients.WaitForTGWAttachmentState(ctx, client, tgwAttID, hvnID, loc, tgwAttState)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	link := newLink(tgwAtt.Location, TgwAttachmentResourceType, tgwAtt.ID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	if err := setTGWAttachmentResourceData(d, tgwAtt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
