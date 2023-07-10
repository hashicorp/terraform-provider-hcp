// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourcePackerRunTask() *schema.Resource {
	return &schema.Resource{
		Description: "The Packer Run Task data source gets the configuration information needed to set up an HCP Packer Registry's run task.",
		ReadContext: dataSourcePackerRunTaskRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPackerTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the HCP Packer Registry is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			// Computed Values
			"organization_id": {
				Description: "The ID of the HCP organization where this channel is located. Always the same as the associated channel.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"endpoint_url": {
				Description: "A unique HCP Packer URL, specific to your HCP organization and HCP Packer registry. The Terraform Cloud run task will send a payload to this URL for image validation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hmac_key": {
				Description: "A secret key that lets HCP Packer verify the run task request.",
				Sensitive:   true,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePackerRunTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(loc.ProjectID)

	resp, err := clients.GetRunTask(ctx, client, loc)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint_url", resp.APIURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hmac_key", resp.HmacKey); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
