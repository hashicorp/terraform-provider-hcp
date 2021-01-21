package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster data source provides information about an existing HCP Consul cluster",
		ReadContext: dataSourceConsulClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterTimeoutDuration,
		},
	}
}

func dataSourceConsulClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
