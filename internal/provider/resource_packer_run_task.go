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

func resourcePackerRunTask() *schema.Resource {
	return &schema.Resource{
		Description:   "The Packer Run Task resource allows you to regenerate the HMAC key for an HCP Packer Registry's run task.",
		CreateContext: resourcePackerRunTaskCreate,
		ReadContext:   resourcePackerRunTaskRead,
		UpdateContext: resourcePackerRunTaskUpdate,
		DeleteContext: resourcePackerRunTaskDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultPackerTimeout,
			Default: &defaultPackerTimeout,
			Update:  &defaultPackerTimeout,
			Delete:  &defaultPackerTimeout,
		},
		CustomizeDiff: resourcePackerRunTaskCustomizeDiff,
		Schema: map[string]*schema.Schema{
			// Optional inputs
			"regenerate_hmac": {
				Description: "If true, the HMAC Key (`hmac_key`) will be regenerated during `terraform apply` and the resource will always cause a non-empty plan. Changing `regenerate_hmac` to false (or removing it from the config) should not result in a plan.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
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

func resourcePackerRunTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := clients.GetRunTask(ctx, client, loc)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(loc.ProjectID)
	if err := d.Set("endpoint_url", resp.APIURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hmac_key", resp.HmacKey); err != nil {
		return diag.FromErr(err)
	}

	regenerateHmac, ok := d.GetOk("regenerate_hmac")
	if ok && regenerateHmac.(bool) {
		resp, err := clients.RegenerateHMAC(ctx, client, loc)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("hmac_key", resp.HmacKey); err != nil {
			return diag.FromErr(err)
		}
		// Always set `regenerate_hmac` to false so that it can be set to false by
		// the user without generating a diff.
		if err := d.Set("regenerate_hmac", false); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourcePackerRunTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourcePackerRunTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	loc, err := getAndUpdateLocationResourceData(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(loc.ProjectID)

	regenerateHmac, ok := d.GetOk("regenerate_hmac")
	if ok && regenerateHmac.(bool) {
		resp, err := clients.RegenerateHMAC(ctx, client, loc)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("hmac_key", resp.HmacKey); err != nil {
			return diag.FromErr(err)
		}
		// Always set `regenerate_hmac` to false so that it can be set to false by
		// the user without generating a diff.
		if err := d.Set("regenerate_hmac", false); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourcePackerRunTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This is a no-op
	return nil
}

func resourcePackerRunTaskCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if d.Get("regenerate_hmac").(bool) {
		if err := d.SetNewComputed("hmac_key"); err != nil {
			return err
		}
	}
	return nil
}
