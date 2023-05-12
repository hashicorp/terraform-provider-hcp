// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var projectDefaultTimeout = time.Minute * 1

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description: "The project data source provides information about an existing project in HCP.",
		ReadContext: dataSourceProjectRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &projectDefaultTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"project_id": {
				Description:  "The ID of the HCP project.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			// Computed outputs
			"name": {
				Description: "The name of the HCP project.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The description of the HCP project.",
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
			"updated_at": {
				Description: "The time that the HCP project was updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceProjectRead is the func to implement reading of an
// HCP project.
func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	// Check for an existing project
	log.Printf("[INFO] Reading project (%s)", projectID)
	project, err := clients.GetProjectByID(ctx, client, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)

	if err := setProjectResourceData(d, project); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setProjectResourceData(d *schema.ResourceData, project *models.ResourcemanagerProject) error {
	if err := d.Set("project_id", project.ID); err != nil {
		return err
	}
	if err := d.Set("name", project.Name); err != nil {
		return err
	}
	if err := d.Set("created_at", project.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("state", project.State); err != nil {
		return err
	}
	if err := d.Set("organization_id", project.Parent.ID); err != nil {
		return err
	}
	return nil
}
