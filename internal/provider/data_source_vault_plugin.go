// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"strings"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceVaultPlugin() *schema.Resource {
	return &schema.Resource{
		Description: "The Vault plugin data source provides information about an existing HCP Vault plugin",
		ReadContext: dataSourceVaultPluginRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultVaultPluginTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Vault cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"plugin_name": {
				Description: "The name of the plugin - Valid options for plugin name - 'venafi-pki-backend'",
				Type:        schema.TypeString,
				Required:    true,
			},
			"plugin_type": {
				Description:      "The type of the plugin - Valid options for plugin type - 'SECRET', 'AUTH', 'DATABASE'",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateVaultPluginType,
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the HCP Vault cluster is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsUUID,
				Computed:     true,
			},
		},
	}
}

func dataSourceVaultPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterID := d.Get("cluster_id").(string)
	pluginName := d.Get("plugin_name").(string)
	pluginTypeString := d.Get("plugin_type").(string)
	pluginType := vaultmodels.HashicorpCloudVault20201125PluginType(pluginTypeString)
	client := meta.(*clients.Client)

	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] Listing plugins for Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	pluginsResp, err := clients.ListPlugins(ctx, client, loc, clusterID)
	if err != nil {
		log.Printf("[ERROR] Vault cluster (%s) failed to list plugins", clusterID)
		return diag.FromErr(err)
	}

	found := false
	for _, plugin := range pluginsResp.Plugins {
		if strings.EqualFold(pluginName, plugin.PluginName) && pluginType == *plugin.PluginType && plugin.IsRegistered {
			found = true
			d.SetId(vaultPluginResourceID(projectID, clusterID, pluginTypeString, pluginName))
			break
		}
	}

	// If Plugin found, update resource data.
	if found {
		if err := setVaultPluginResourceData(d, projectID, clusterID, pluginName, pluginTypeString); err != nil {
			return diag.FromErr(err)
		}
		return nil
	} else {
		return diag.Errorf("unable to retrieve registered plugin: %s", pluginName)
	}
}
