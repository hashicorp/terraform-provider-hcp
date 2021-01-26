package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHvn() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN data source provides information about an existing HashiCorp Virtual Network.",
		ReadContext: dataSourceHvnRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &hvnDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"hvn_id": {
				Description:      "The ID of the HashiCorp Virtual Network.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"cloud_provider": {
				Description:      "The provider where the HVN is located.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringInSlice(hvnResourceCloudProviders, true),
			},
			"region": {
				Description: "The region where the HVN is located.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Computed outputs
			"cidr_block": {
				Description: "The CIDR range of the HVN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
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
			"created_at": {
				Description: "The time that the HVN was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceHvnRead is the func to implement reading of an
// HashiCorp Virtual Network (HVN)
func dataSourceHvnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
