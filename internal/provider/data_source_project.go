// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectDefaultTimeout = time.Minute * 1

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description: "The project data source provides information about an existing HashiCorp project.",
		ReadContext: dataSourceProjectRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &projectDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"project_id": {
				Description:      "The ID of the HCP project.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed outputs
			"name": {
				Description: "The name of the HCP project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the HCP project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization_id": {
				Description: "The ID of the HCP organization the project belongs to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the HCP project was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceProjectRead is the func to implement reading of an
// HCP project.
func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
