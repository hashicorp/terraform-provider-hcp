// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a vault plugin operation should timeout.
var defaultVaultPluginTimeout = time.Minute * 1

func resourceVaultPlugin() *schema.Resource {
	return &schema.Resource{
		Description:   "The Vault plugin resource allows you to manage an HCP Vault plugin.",
		CreateContext: resourceVaultPluginCreate,
		ReadContext:   resourceVaultPluginRead,
		UpdateContext: resourceVaultPluginUpdate,
		DeleteContext: resourceVaultPluginDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultVaultPluginTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVaultPluginImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Vault cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"plugin_name": {
				Description: "The name of the plugin - Valid options for plugin name - 'venafi-pki-backend'",
				Type:        schema.TypeString,
				Required:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"plugin_type": {
				Description:      "The type of the plugin - Valid options for plugin type - 'SECRET', 'AUTH', 'DATABASE'",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateVaultPluginType,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			// Optional inputs
			"project_id": {
				Description: `
The ID of the HCP project where the HCP Vault cluster is located. 
If not specified, the project specified in the HCP Provider config block will be used, if configured.
If a project is not configured in the HCP Provider config block, the oldest project in the organization will be used.`,
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Computed:     true,
			},
		},
	}
}

func resourceVaultPluginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}
	pluginName := d.Get("plugin_name").(string)
	pluginType := d.Get("plugin_type").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] Adding Vault Plugin (%s) on Vault Cluster (%s)", pluginName, clusterID)

	req := &vaultmodels.HashicorpCloudVault20201125AddPluginRequest{PluginName: pluginName, PluginType: pluginType}
	_, err = clients.AddPlugin(ctx, client, loc, clusterID, req)
	if err != nil {
		return diag.Errorf("error adding plugin (%s) to Vault cluster (%s): %v", pluginName, clusterID, err)
	}

	d.SetId(vaultPluginResourceID(projectID, clusterID, pluginType, pluginName))

	if err := setVaultPluginResourceData(d, projectID, clusterID, pluginName, pluginType); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVaultPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	idParts := strings.SplitN(d.Id(), "/", 8)

	clusterID := idParts[4]
	projectID := idParts[2]
	pluginName := idParts[7]
	pluginTypeString := idParts[6]
	pluginType := vaultmodels.HashicorpCloudVault20201125PluginType(pluginTypeString)

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

	for _, plugin := range pluginsResp.Plugins {
		if strings.EqualFold(pluginName, plugin.PluginName) && pluginType == *plugin.PluginType && plugin.IsRegistered {
			// Cluster found, update resource data.
			if err := setVaultPluginResourceData(d, loc.ProjectID, clusterID, pluginName, pluginTypeString); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}

	// if plugin is not registered, remove from state
	d.SetId("")
	return nil
}

func resourceVaultPluginUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)
	clusterID := d.Get("cluster_id").(string)
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	pluginName := d.Get("plugin_name").(string)
	pluginType := d.Get("plugin_type").(string)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	if d.HasChanges("cluster_id", "project_id", "plugin_name", "plugin_type") {

		oldPlugin, _ := d.GetChange("vault_plugin")

		config, ok := oldPlugin.(map[string]interface{})
		if !ok {
			return diag.Errorf("could not parse old plugin config: %v", err)
		}

		oldPluginName := config["plugin_name"].(string)
		oldPluginType := config["plugin_type"].(string)
		oldProjectID, err := GetProjectID(config["project_id"].(string), client.Config.ProjectID)
		if err != nil {
			return diag.Errorf("unable to retrieve project ID: %v", err)
		}
		oldClusterID := config["cluster_id"].(string)

		oldLoc := &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: client.Config.OrganizationID,
			ProjectID:      oldProjectID,
		}

		log.Printf("[INFO] Deleting Vault Plugin (%s) on Vault Cluster (%s)", pluginName, clusterID)
		req := &vaultmodels.HashicorpCloudVault20201125DeletePluginRequest{PluginName: oldPluginName, PluginType: oldPluginType}
		_, err = clients.DeletePlugin(ctx, client, oldLoc, clusterID, req)
		if err != nil {
			return diag.Errorf("error deleting plugin (%s) on Vault cluster (%s): %v", oldPluginName, oldClusterID, err)
		}
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] Adding Vault Plugin (%s) on Vault Cluster (%s)", pluginName, clusterID)
	req := &vaultmodels.HashicorpCloudVault20201125AddPluginRequest{PluginName: pluginName, PluginType: pluginType}
	_, err = clients.AddPlugin(ctx, client, loc, clusterID, req)
	if err != nil {
		return diag.Errorf("error adding plugin (%s) to Vault cluster (%s): %v", pluginName, clusterID, err)
	}

	if err := setVaultPluginResourceData(d, projectID, clusterID, pluginName, pluginType); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVaultPluginDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	projectID, err := GetProjectID(d.Get("project_id").(string), client.Config.ProjectID)
	pluginName := d.Get("plugin_name").(string)
	pluginType := d.Get("plugin_type").(string)
	if err != nil {
		return diag.Errorf("unable to retrieve project ID: %v", err)
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	log.Printf("[INFO] Deleting Vault Plugin (%s) on Vault Cluster (%s)", pluginName, clusterID)

	req := &vaultmodels.HashicorpCloudVault20201125DeletePluginRequest{PluginName: pluginName, PluginType: pluginType}
	_, err = clients.DeletePlugin(ctx, client, loc, clusterID, req)
	if err != nil {
		return diag.Errorf("error deleting plugin (%s) on Vault cluster (%s): %v", pluginName, pluginType, err)
	}

	return nil
}

// setVaultPluginResourceData sets the KV pairs of the Vault cluster resource schema.
func setVaultPluginResourceData(d *schema.ResourceData, projectID string, clusterID string, pluginName string, pluginType string) error {
	if err := d.Set("cluster_id", clusterID); err != nil {
		return err
	}

	if err := d.Set("project_id", projectID); err != nil {
		return err
	}

	if err := d.Set("plugin_name", pluginName); err != nil {
		return err
	}

	if err := d.Set("plugin_type", pluginType); err != nil {
		return err
	}

	return nil
}

// resourceHVNRouteImport implements the logic necessary to import an
// un-tracked (by Terraform) HVN route resource into Terraform state.
func resourceVaultPluginImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// with multi-projects, import arguments must become dynamic:
	// use explicit project ID with terraform import:
	//   terraform import hcp_vault_plugin.test {project_id}:{cluster_id}:{plugin_type}:{plugin_name}
	// use default project ID from provider:
	//   terraform import hcp_vault_plugin.test {cluster_id}:{plugin_type}:{plugin_name}

	client := meta.(*clients.Client)
	projectID := ""
	clusterID := ""
	pluginType := ""
	pluginName := ""
	var err error

	idParts := strings.SplitN(d.Id(), ":", 4)
	if len(idParts) == 4 { // {project_id}:{cluster_id}:{plugin_type}:{plugin_name}
		if idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {project_id}:{cluster_id}:{plugin_type}:{plugin_name}", d.Id())
		}
		projectID = idParts[0]
		clusterID = idParts[1]
		pluginType = idParts[2]
		pluginName = idParts[3]
	} else if len(idParts) == 3 { // {cluster_id}:{plugin_type}:{plugin_name}
		if idParts[0] == "" || idParts[1] == "" {
			return nil, fmt.Errorf("unexpected format of ID (%q), expected {cluster_id}:{plugin_type}:{plugin_name}", d.Id())
		}
		projectID, err = GetProjectID(projectID, client.Config.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve project ID: %v", err)
		}
		clusterID = idParts[0]
		pluginType = idParts[1]
		pluginName = idParts[2]
	} else {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected {cluster_id}:{plugin_type}:{plugin_name} or {project_id}:{cluster_id}:{plugin_type}:{plugin_name}", d.Id())
	}

	d.SetId(vaultPluginResourceID(projectID, clusterID, pluginType, pluginName))

	return []*schema.ResourceData{d}, nil
}

func vaultPluginResourceID(projectID string, clusterID string, pluginType string, pluginName string) string {
	return fmt.Sprintf("/project/%s/%s/%s/plugin/%s/%s",
		projectID,
		VaultClusterResourceType,
		clusterID,
		pluginType,
		pluginName)
}
