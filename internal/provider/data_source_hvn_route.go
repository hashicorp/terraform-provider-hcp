package provider

import (
	"context"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceHVNRoute() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN Route data source provides information about an existing HVN Route.",
		ReadContext: dataSourceHVNRouteRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnRouteDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network (HVN).",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"destination_cidr": {
				Description: "The destination CIDR of the HVN Route",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"organization_id": {
				Description: "The ID of the HCP organization where the HVN Route is located. Always matches the HVN's organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the HCP project where the HVN Route is located. Always matches the HVN's project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceHVNRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	hvnID := d.Get("hvn_id").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	_, err := clients.ListHVNRoutes(ctx, client, hvnID, loc)
	if err != nil {
		return diag.FromErr(err)
	}

	// if err := setHVNRouteResourceData(d, routes); err != nil {
	// 	return diag.FromErr(err)
	// }

	return nil
}
