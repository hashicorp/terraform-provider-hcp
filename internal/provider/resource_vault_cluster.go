// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	vaultmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-service/stable/2020-11-25/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/input"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultVaultClusterTimeout = time.Minute * 5

// createUpdateVaultClusterTimeout is the amount of time that can elapse
// before a cluster create operation should timeout.
var createUpdateVaultClusterTimeout = time.Minute * 75

// deleteVaultClusterTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteVaultClusterTimeout = time.Minute * 75

func resourceVaultCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The Vault cluster resource allows you to manage an HCP Vault cluster.",
		CreateContext: resourceVaultClusterCreate,
		ReadContext:   resourceVaultClusterRead,
		UpdateContext: resourceVaultClusterUpdate,
		DeleteContext: resourceVaultClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  &createUpdateVaultClusterTimeout,
			Update:  &createUpdateVaultClusterTimeout,
			Delete:  &deleteVaultClusterTimeout,
			Default: &defaultVaultClusterTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceVaultClusterImport,
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
			"hvn_id": {
				Description:      "The ID of the HVN this HCP Vault cluster is associated to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional fields
			"tier": {
				Description:      "Tier of the HCP Vault cluster. Valid options for tiers - `dev`, `starter_small`, `standard_small`, `standard_medium`, `standard_large`, `plus_small`, `plus_medium`, `plus_large`. See [pricing information](https://www.hashicorp.com/products/vault/pricing). Changing a cluster's size or tier is only available to admins. See [Scale a cluster](https://registry.terraform.io/providers/hashicorp/hcp/latest/docs/guides/vault-scaling).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateVaultClusterTier,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint. Defaults to false.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"min_vault_version": {
				Description:      "The minimum Vault version to use when creating the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSemVer,
				ForceNew:         true,
			},
			// Only applies to Plus tier HCP Vault clusters
			"primary_link": {
				Description: "The `self_link` of the HCP Vault Plus tier cluster which is the primary in the performance replication setup with this HCP Vault Plus tier cluster. If not specified, it is a standalone Plus tier HCP Vault cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"paths_filter": {
				Description: "The performance replication [paths filter](https://developer.hashicorp.com/vault/tutorials/cloud-ops/vault-replication-terraform). Applies to performance replication secondaries only and operates in \"deny\" mode only.",
				Type:        schema.TypeList,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validateVaultPathsFilter,
				},
				Optional: true,
			},
			"organization_id": {
				Description: "The ID of the organization this HCP Vault cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Vault cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_provider": {
				Description: "The provider where the HCP Vault cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "The region where the HCP Vault cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace": {
				Description: "The name of the customer namespace this HCP Vault cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"metrics_config": {
				Description: "The metrics configuration for export. (https://developer.hashicorp.com/vault/tutorials/cloud-monitoring/vault-metrics-guide#metrics-streaming-configuration)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grafana_endpoint": {
							Description: "Grafana endpoint for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"grafana_user": {
							Description: "Grafana user for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"grafana_password": {
							Description: "Grafana password for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"splunk_hecendpoint": {
							Description: "Splunk endpoint for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"splunk_token": {
							Description: "Splunk token for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"datadog_api_key": {
							Description: "Datadog api key for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"datadog_region": {
							Description: "Datadog region for streaming metrics",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"audit_log_config": {
				Description: "The audit logs configuration for export. (https://developer.hashicorp.com/vault/tutorials/cloud-monitoring/vault-metrics-guide#metrics-streaming-configuration)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grafana_endpoint": {
							Description: "Grafana endpoint for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"grafana_user": {
							Description: "Grafana user for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"grafana_password": {
							Description: "Grafana password for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"splunk_hecendpoint": {
							Description: "Splunk endpoint for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"splunk_token": {
							Description: "Splunk token for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"datadog_api_key": {
							Description: "Datadog api key for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"datadog_region": {
							Description: "Datadog region for streaming audit logs",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"vault_version": {
				Description: "The Vault version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"major_version_upgrade_config": {
				Description: "The Major Version Upgrade configuration.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"upgrade_type": {
							Description:      "The major upgrade type for the cluster. Valid options for upgrade type - `AUTOMATIC`, `SCHEDULED`, `MANUAL`",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validateVaultUpgradeType,
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
						},
						"maintenance_window_day": {
							Description:      "The maintenance day of the week for scheduled upgrades. Valid options for maintenance window day - `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`",
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateVaultUpgradeWindowDay,
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
						},
						"maintenance_window_time": {
							Description:      "The maintenance time frame for scheduled upgrades. Valid options for maintenance window time - `WINDOW_12AM_4AM`, `WINDOW_6AM_10AM`, `WINDOW_12PM_4PM`, `WINDOW_6PM_10PM`",
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateVaultUpgradeWindowTime,
							DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
						},
					},
				},
			},
			"vault_public_endpoint_url": {
				Description: "The public URL for the Vault cluster. This will be empty if `public_endpoint` is `false`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vault_private_endpoint_url": {
				Description: "The private URL for the Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "The time that the Vault cluster was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVaultClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)
	hvnID := d.Get("hvn_id").(string)
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	// Get metrics audit config and MVU config first so we can validate and fail faster.
	metricsConfig, diagErr := getObservabilityConfig("metrics_config", d)
	if diagErr != nil {
		return diagErr
	}
	auditConfig, diagErr := getObservabilityConfig("audit_log_config", d)
	if diagErr != nil {
		return diagErr
	}
	mvuConfig, error := getMajorVersionUpgradeConfig(d)
	if diagErr != nil {
		return error
	}

	// Use the hvn to get provider and region.
	hvn, err := clients.GetHvnByID(ctx, client, loc, hvnID)
	if err != nil {
		return diag.Errorf("unable to find existing HVN (%s): %v", hvnID, err)
	}
	loc.Region = &sharedmodels.HashicorpCloudLocationRegion{
		Provider: hvn.Location.Region.Provider,
		Region:   hvn.Location.Region.Region,
	}

	// Check for an existing Vault cluster.
	_, err = clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Vault cluster (%s): %v", clusterID, err)
		}

		// A 404 indicates a Vault cluster was not found.
		log.Printf("[INFO] Vault cluster (%s) not found, proceeding with create", clusterID)
	} else {
		return diag.Errorf("a Vault cluster with cluster_id=%q in project_id=%q already exists - to be managed via Terraform this resource needs to be imported into the State.  Please see the resource documentation for hcp_vault_cluster for more information.", clusterID, loc.ProjectID)
	}

	// If no min_vault_version is set, an empty version is passed and the backend will set a default version.
	var vaultVersion string
	v, ok := d.GetOk("min_vault_version")
	if ok {
		vaultVersion = input.NormalizeVersion(v.(string))
	}

	publicEndpoint := d.Get("public_endpoint").(bool)

	// If the cluster has a primary_link, make sure the link is valid
	diagErr, primaryClusterModel := validatePerformanceReplicationChecksAndReturnPrimaryIfAny(ctx, client, d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[INFO] Creating Vault cluster (%s)", clusterID)

	var vaultCluster *vaultmodels.HashicorpCloudVault20201125InputCluster
	if getPrimaryLinkIfAny(d) != "" {
		primaryClusterLink := newLink(primaryClusterModel.Location, VaultClusterResourceType, primaryClusterModel.ID)
		var pathsFilter *vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter
		mode := vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilterModeDENY
		if paths, ok := d.GetOk("paths_filter"); ok {
			pathStrings := getPathStrings(paths)
			pathsFilter = &vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter{
				Mode:  &mode,
				Paths: pathStrings,
			}
		}
		vaultCluster = &vaultmodels.HashicorpCloudVault20201125InputCluster{
			Config: &vaultmodels.HashicorpCloudVault20201125InputClusterConfig{
				VaultConfig: &vaultmodels.HashicorpCloudVault20201125VaultConfig{
					// Secondary clusters inherit InitialVersion from their primary's current version
					InitialVersion: primaryClusterModel.CurrentVersion,
				},
				Tier: primaryClusterModel.Config.Tier,
				NetworkConfig: &vaultmodels.HashicorpCloudVault20201125InputNetworkConfig{
					NetworkID:        hvn.ID,
					PublicIpsEnabled: publicEndpoint,
				},
			},
			ID:                                   clusterID,
			Location:                             loc,
			PerformanceReplicationPrimaryCluster: primaryClusterLink,
			PerformanceReplicationPathsFilter:    pathsFilter,
		}
	} else {
		if _, ok := d.GetOk("paths_filter"); ok {
			return diag.Errorf("only performance replication secondaries may specify a paths_filter")
		}

		var tier *vaultmodels.HashicorpCloudVault20201125Tier
		t, ok := d.GetOk("tier")
		if ok {
			tier = vaultmodels.HashicorpCloudVault20201125Tier(strings.ToUpper(t.(string))).Pointer()
		}

		vaultCluster = &vaultmodels.HashicorpCloudVault20201125InputCluster{
			Config: &vaultmodels.HashicorpCloudVault20201125InputClusterConfig{
				VaultConfig: &vaultmodels.HashicorpCloudVault20201125VaultConfig{
					InitialVersion: vaultVersion,
				},
				Tier: tier,
				NetworkConfig: &vaultmodels.HashicorpCloudVault20201125InputNetworkConfig{
					NetworkID:        hvn.ID,
					PublicIpsEnabled: publicEndpoint,
				},
			},
			ID:       clusterID,
			Location: loc,
		}
	}

	if metricsConfig != nil {
		vaultCluster.Config.MetricsConfig = metricsConfig
	}
	if auditConfig != nil {
		vaultCluster.Config.AuditLogExportConfig = auditConfig
	}

	payload, err := clients.CreateVaultCluster(ctx, client, loc, vaultCluster)
	if err != nil {
		return diag.Errorf("unable to create Vault cluster (%s): %v", clusterID, err)
	}

	link := newLink(loc, VaultClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(url)

	// Wait for the Vault cluster to be created.
	if err := clients.WaitForOperation(ctx, client, "create Vault cluster", loc, payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Vault cluster (%s): %v", payload.ClusterID, err)
	}

	log.Printf("[INFO] Created Vault cluster (%s)", payload.ClusterID)

	// Get the created Vault cluster.
	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, payload.ClusterID)

	if err != nil {
		return diag.Errorf("unable to retrieve Vault cluster (%s): %v", payload.ClusterID, err)
	}

	// If we pass the major version upgrade configuration we need to update it after the creation of the cluster,
	// since the cluster is created by default to automatic upgrade
	if mvuConfig != nil {
		_, err := clients.UpdateVaultMajorVersionUpgradeConfig(ctx, client, cluster.Location, payload.ClusterID, mvuConfig)
		if err != nil {
			return diag.Errorf("error updating Vault cluster major version upgrade config (%s): %v", payload.ClusterID, err)
		}

		// refresh the created Vault cluster.
		cluster, err = clients.GetVaultClusterByID(ctx, client, loc, payload.ClusterID)
		if err != nil {
			return diag.Errorf("unable to retrieve Vault cluster (%s): %v", payload.ClusterID, err)
		}
	}

	if err := setVaultClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVaultClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), VaultClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Vault cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Vault cluster (%s): %v", clusterID, err)
	}

	// The Vault cluster was already deleted, remove from state.
	if *cluster.State == vaultmodels.HashicorpCloudVault20201125ClusterStateDELETED {
		log.Printf("[WARN] Vault cluster (%s) failed to provision, removing from state", clusterID)
		d.SetId("")
		return nil
	}

	// Cluster found, update resource data.
	if err := setVaultClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVaultClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), VaultClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Vault cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Vault cluster (%s): %v", clusterID, err)
	}

	// Confirm at least one modifiable field has changed
	if !d.HasChanges("tier", "public_endpoint", "paths_filter", "metrics_config", "audit_log_config", "major_version_upgrade_config") {
		return nil
	}

	// Get metrics audit config and mvu config first so we can validate and fail faster
	mvuConfig, diagErr := getMajorVersionUpgradeConfig(d)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("tier") || d.HasChange("metrics_config") || d.HasChange("audit_log_config") {
		diagErr := updateVaultClusterConfig(ctx, client, d, cluster, clusterID)
		if diagErr != nil {
			return diagErr
		}
	}

	if d.HasChange("public_endpoint") {
		// Invoke update public IPs endpoint.
		updateResp, err := clients.UpdateVaultClusterPublicIps(ctx, client, cluster.Location, clusterID, d.Get("public_endpoint").(bool))
		if err != nil {
			return diag.Errorf("error updating Vault cluster public endpoint (%s): %v", clusterID, err)
		}

		// Wait for the update cluster operation.
		if err := clients.WaitForOperation(ctx, client, "update Vault cluster public endpoint", cluster.Location, updateResp.Operation.ID); err != nil {
			return diag.Errorf("unable to update Vault cluster public endpoint (%s): %v", clusterID, err)
		}
	}

	if d.HasChange("paths_filter") {
		if paths, ok := d.GetOk("paths_filter"); ok {
			// paths_filter is present. Check that it is a secondary, then update.
			if _, ok := d.GetOk("primary_link"); !ok {
				return diag.Errorf("only performance replication secondaries may specify a paths_filter")
			}

			// Invoke update paths filter endpoint.
			pathStrings := getPathStrings(paths)
			mode := vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilterModeDENY
			updateResp, err := clients.UpdateVaultPathsFilter(ctx, client, cluster.Location, clusterID, vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter{
				Mode:  &mode,
				Paths: pathStrings,
			})
			if err != nil {
				return diag.Errorf("error updating Vault cluster paths filter (%s): %v", clusterID, err)
			}

			// Wait for the update paths filter operation.
			if err := clients.WaitForOperation(ctx, client, "update Vault cluster paths filter", cluster.Location, updateResp.Operation.ID); err != nil {
				return diag.Errorf("unable to update Vault cluster paths filter (%s): %v", clusterID, err)
			}
		} else {
			// paths_filter is not present. Delete the paths_filter.
			deleteResp, err := clients.DeleteVaultPathsFilter(ctx, client, cluster.Location, clusterID)
			if err != nil {
				return diag.Errorf("error deleting Vault cluster paths filter (%s): %v", clusterID, err)
			}

			// Wait for the delete paths filter operation.
			if err := clients.WaitForOperation(ctx, client, "delete Vault cluster paths filter", cluster.Location, deleteResp.Operation.ID); err != nil {
				return diag.Errorf("unable to delete Vault cluster paths filter (%s): %v", clusterID, err)
			}
		}
	}

	if mvuConfig != nil {
		_, err := clients.UpdateVaultMajorVersionUpgradeConfig(ctx, client, cluster.Location, clusterID, mvuConfig)
		if err != nil {
			return diag.Errorf("error updating Vault cluster major version upgrade config (%s): %v", clusterID, err)
		}
	}

	// Get the updated Vault cluster.
	cluster, err = clients.GetVaultClusterByID(ctx, client, loc, clusterID)

	if err != nil {
		return diag.Errorf("unable to retrieve Vault cluster (%s): %v", clusterID, err)
	}

	if err := setVaultClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVaultClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), VaultClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Deleting Vault cluster (%s)", clusterID)

	deleteResp, err := clients.DeleteVaultCluster(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Vault cluster (%s) not found, so no action was taken", clusterID)
			return nil
		}

		return diag.Errorf("unable to delete Vault cluster (%s): %v", clusterID, err)
	}

	// Wait for the delete cluster operation.
	if err := clients.WaitForOperation(ctx, client, "delete Vault cluster", loc, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Vault cluster (%s): %v", clusterID, err)
	}

	return nil
}
func updateVaultClusterConfig(ctx context.Context, client *clients.Client, d *schema.ResourceData, cluster *vaultmodels.HashicorpCloudVault20201125Cluster, clusterID string) diag.Diagnostics {
	metricsConfig, diagErr := getObservabilityConfig("metrics_config", d)
	if diagErr != nil {
		return diagErr
	}
	auditConfig, diagErr := getObservabilityConfig("audit_log_config", d)
	if diagErr != nil {
		return diagErr
	}
	isSecondary := false
	destTier := getClusterTier(d)
	if d.HasChange("tier") {
		if inPlusTier(string(*cluster.Config.Tier)) {
			// Plus tier clusters scale as a group via the primary cluster.
			// However, it is still worth individually tracking the tier of each cluster so that the
			// provider has the same information as the portal UI and can detect a scaling operation that
			// fails part way through to enable retries.
			// Because the clusters scale as group,
			//  a) replicated clusters may have already scaled due to another resource's update
			//  b) all scaling requests are routed through the primary
			// It is important to keep the tier of all replicated clusters in sync.

			// Because of (a), check that the scaling operation is necessary.
			// If the cluster has the same tier but the metrics/audit_log changed, we want to update the cluster anyway to change the info.
			if *cluster.Config.Tier == vaultmodels.HashicorpCloudVault20201125Tier(*destTier) && !d.HasChange("metrics_config") && !d.HasChange("audit_log_config") {
				return nil
			} else {
				printPlusScalingWarningMsg()
				primaryLink := getPrimaryLinkIfAny(d)
				if primaryLink != "" {
					// Because of (b), if the cluster is a secondary, issue the actual API request to the primary.
					isSecondary = true
					if d.HasChange("metrics_config") || d.HasChange("audit_log_config") {
						updateResp, err := clients.UpdateVaultClusterConfig(ctx, client, cluster.Location, cluster.ID, destTier, metricsConfig, auditConfig)
						if err != nil {
							return diag.Errorf("error updating Vault cluster (%s): %v", clusterID, err)
						}

						// Wait for the update cluster operation.
						if err := clients.WaitForOperation(ctx, client, "update Vault cluster", cluster.Location, updateResp.Operation.ID); err != nil {
							return diag.Errorf("unable to update Vault cluster (%s): %v", clusterID, err)
						}
					}
					var getPrimaryErr diag.Diagnostics
					cluster, getPrimaryErr = getPrimaryClusterFromLink(ctx, client, primaryLink)
					if getPrimaryErr != nil {
						return getPrimaryErr
					}
				}
			}
		}
	}

	// If is secondary since we're scaling via the primary we don't update the primary metrics/auditLog.
	if isSecondary {
		metricsConfig = nil
		auditConfig = nil
	}
	// Invoke update endpoint.
	updateResp, err := clients.UpdateVaultClusterConfig(ctx, client, cluster.Location, cluster.ID, destTier, metricsConfig, auditConfig)
	if err != nil {
		return diag.Errorf("error updating Vault cluster (%s): %v", clusterID, err)
	}
	// Wait for the update cluster operation.
	if err := clients.WaitForOperation(ctx, client, "update Vault cluster", cluster.Location, updateResp.Operation.ID); err != nil {
		return diag.Errorf("unable to update Vault cluster (%s): %v", clusterID, err)
	}

	return nil
}

func getClusterTier(d *schema.ResourceData) *string {
	// If we don't change the tier, return nil so we don't pass the tier to the update.
	if d.HasChange("tier") {
		tier := strings.ToUpper(d.Get("tier").(string))
		return &tier
	}
	return nil
}

// setVaultClusterResourceData sets the KV pairs of the Vault cluster resource schema.
func setVaultClusterResourceData(d *schema.ResourceData, cluster *vaultmodels.HashicorpCloudVault20201125Cluster) error {

	if err := d.Set("cluster_id", cluster.ID); err != nil {
		return err
	}

	if err := d.Set("hvn_id", cluster.Config.NetworkConfig.NetworkID); err != nil {
		return err
	}

	if err := d.Set("organization_id", cluster.Location.OrganizationID); err != nil {
		return err
	}

	if err := d.Set("project_id", cluster.Location.ProjectID); err != nil {
		return err
	}

	if err := d.Set("cloud_provider", cluster.Location.Region.Provider); err != nil {
		return err
	}

	if err := d.Set("region", cluster.Location.Region.Region); err != nil {
		return err
	}

	if err := d.Set("tier", cluster.Config.Tier); err != nil {
		return err
	}

	if err := d.Set("vault_version", cluster.CurrentVersion); err != nil {
		return err
	}

	if err := d.Set("namespace", cluster.Config.VaultConfig.Namespace); err != nil {
		return err
	}

	if err := d.Set("state", cluster.State); err != nil {
		return err
	}

	publicEndpoint := cluster.Config.NetworkConfig.PublicIpsEnabled
	if err := d.Set("public_endpoint", publicEndpoint); err != nil {
		return err
	}

	if err := d.Set("metrics_config", flattenObservabilityConfig(cluster.Config.MetricsConfig, d, "metrics_config")); err != nil {
		return err
	}

	if err := d.Set("audit_log_config", flattenObservabilityConfig(cluster.Config.AuditLogExportConfig, d, "audit_log_config")); err != nil {
		return err
	}

	if err := d.Set("major_version_upgrade_config", flattenMajorVersionUpgradeConfig(cluster.Config.MajorVersionUpgradeConfig, d)); err != nil {
		return err
	}

	if publicEndpoint {
		// Port 8200 required to communicate with HCP Vault via HTTPS
		if err := d.Set("vault_public_endpoint_url", fmt.Sprintf("https://%s:8200", cluster.DNSNames.Public)); err != nil {
			return err
		}
	}

	// Port 8200 required to communicate with HCP Vault via HTTPS
	if err := d.Set("vault_private_endpoint_url", fmt.Sprintf("https://%s:8200", cluster.DNSNames.Private)); err != nil {
		return err
	}

	if err := d.Set("created_at", cluster.CreatedAt.String()); err != nil {
		return err
	}

	link := newLink(cluster.Location, VaultClusterResourceType, cluster.ID)
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	if cluster.PerformanceReplicationInfo != nil {
		prInfo := cluster.PerformanceReplicationInfo
		if prInfo.PrimaryClusterLink != nil {
			primaryLink, err := linkURL(cluster.PerformanceReplicationInfo.PrimaryClusterLink)
			if err != nil {
				return err
			}
			if err := d.Set("primary_link", primaryLink); err != nil {
				return err
			}
		}

		if prInfo.PathsFilter != nil && prInfo.PathsFilter.Paths != nil {
			if err := d.Set("paths_filter", prInfo.PathsFilter.Paths); err != nil {
				return err
			}
		} else {
			err = d.Set("paths_filter", nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenObservabilityConfig(config *vaultmodels.HashicorpCloudVault20201125ObservabilityConfig, d *schema.ResourceData, propertyName string) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	configMap := map[string]interface{}{}

	if grafana := config.Grafana; grafana != nil {
		configMap["grafana_endpoint"] = grafana.Endpoint
		configMap["grafana_user"] = grafana.User
		// Since the API return this sensitive fields as redacted, we don't update it on the config in this situations
		if grafana.Password != "redacted" {
			configMap["grafana_password"] = grafana.Password
		} else {
			if configParam, ok := d.GetOk(propertyName); ok && len(configParam.([]interface{})) > 0 {
				config := configParam.([]interface{})[0].(map[string]interface{})
				configMap["grafana_password"] = config["grafana_password"].(string)
			}
		}
	}

	if splunk := config.Splunk; splunk != nil {
		configMap["splunk_hecendpoint"] = splunk.HecEndpoint
		// Since the API return this sensitive fields as redacted, we don't update it on the config in this situations
		if splunk.Token != "redacted" {
			configMap["splunk_token"] = splunk.Token
		} else {
			if configParam, ok := d.GetOk(propertyName); ok && len(configParam.([]interface{})) > 0 {
				config := configParam.([]interface{})[0].(map[string]interface{})
				configMap["splunk_token"] = config["splunk_token"].(string)
			}
		}
	}

	if datadog := config.Datadog; datadog != nil {
		configMap["datadog_region"] = datadog.Region
		// Since the API return this sensitive fields as redacted, we don't update it on the config in this situations
		if datadog.APIKey != "redacted" {
			configMap["datadog_api_key"] = datadog.APIKey
		} else {
			if configParam, ok := d.GetOk(propertyName); ok && len(configParam.([]interface{})) > 0 {
				config := configParam.([]interface{})[0].(map[string]interface{})
				configMap["datadog_api_key"] = config["datadog_api_key"].(string)
			}
		}
	}
	return []interface{}{configMap}
}

func getObservabilityConfig(propertyName string, d *schema.ResourceData) (*vaultmodels.HashicorpCloudVault20201125ObservabilityConfig, diag.Diagnostics) {
	if !d.HasChange(propertyName) {
		return nil, nil
	}

	emptyConfig := vaultmodels.HashicorpCloudVault20201125ObservabilityConfig{
		Grafana: &vaultmodels.HashicorpCloudVault20201125Grafana{},
		Splunk:  &vaultmodels.HashicorpCloudVault20201125Splunk{},
		Datadog: &vaultmodels.HashicorpCloudVault20201125Datadog{},
	}

	// If we don't find the property we return the empty object to be updated and delete the configuration.
	configParam, ok := d.GetOk(propertyName)
	if !ok {
		return &emptyConfig, nil
	}
	configIfaceArr, ok := configParam.([]interface{})
	if !ok || len(configIfaceArr) == 0 {
		return &emptyConfig, nil
	}
	config, ok := configIfaceArr[0].(map[string]interface{})
	if !ok {
		return &emptyConfig, nil
	}

	return getValidObservabilityConfig(config)
}

func getValidObservabilityConfig(config map[string]interface{}) (*vaultmodels.HashicorpCloudVault20201125ObservabilityConfig, diag.Diagnostics) {

	observabilityConfig := vaultmodels.HashicorpCloudVault20201125ObservabilityConfig{}

	grafanaEndpoint := config["grafana_endpoint"].(string)
	grafanaUser := config["grafana_user"].(string)
	grafanaPassword := config["grafana_password"].(string)
	splunkEndpoint := config["splunk_hecendpoint"].(string)
	splunkToken := config["splunk_token"].(string)
	datadogAPIKey := config["datadog_api_key"].(string)
	datadogRegion := config["datadog_region"].(string)

	if grafanaEndpoint != "" || grafanaUser != "" || grafanaPassword != "" {
		if grafanaEndpoint == "" || grafanaUser == "" || grafanaPassword == "" {
			return nil, diag.Errorf("grafana configuration is invalid: configuration information missing")
		} else if splunkEndpoint != "" || splunkToken != "" || datadogAPIKey != "" || datadogRegion != "" {
			return nil, diag.Errorf("multiple configurations found: must contain configuration for only one provider")
		}
		observabilityConfig.Grafana = &vaultmodels.HashicorpCloudVault20201125Grafana{
			Endpoint: grafanaEndpoint,
			User:     grafanaUser,
			Password: grafanaPassword,
		}
	}

	if splunkEndpoint != "" || splunkToken != "" {
		if splunkEndpoint == "" || splunkToken == "" {
			return nil, diag.Errorf("splunk configuration is invalid: configuration information missing")
		} else if datadogAPIKey != "" || datadogRegion != "" {
			return nil, diag.Errorf("multiple configurations found: must contain configuration for only one provider")
		}
		observabilityConfig.Splunk = &vaultmodels.HashicorpCloudVault20201125Splunk{
			HecEndpoint: splunkEndpoint,
			Token:       splunkToken,
		}
	}

	if datadogAPIKey != "" || datadogRegion != "" {
		if datadogAPIKey == "" || datadogRegion == "" {
			return nil, diag.Errorf("datadog configuration is invalid: configuration information missing")
		}
		observabilityConfig.Datadog = &vaultmodels.HashicorpCloudVault20201125Datadog{
			APIKey: datadogAPIKey,
			Region: datadogRegion,
		}
	}

	return &observabilityConfig, nil
}

func getMajorVersionUpgradeConfig(d *schema.ResourceData) (*vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfig, diag.Diagnostics) {
	if !d.HasChange("major_version_upgrade_config") {
		return nil, nil
	}
	configParam, ok := d.GetOk("major_version_upgrade_config")
	if !ok {
		return nil, nil
	}

	configIfaceArr, ok := configParam.([]interface{})
	if !ok || len(configIfaceArr) == 0 {
		return nil, nil
	}

	config, ok := configIfaceArr[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	tier := vaultmodels.HashicorpCloudVault20201125TierDEV
	inputTier, ok := d.GetOk("tier")
	if ok {
		tier = vaultmodels.HashicorpCloudVault20201125Tier(strings.ToUpper(inputTier.(string)))
	}

	if !ok || len(configIfaceArr) == 0 {
		return nil, nil
	}

	return getValidMajorVersionUpgradeConfig(config, tier)
}

func getValidMajorVersionUpgradeConfig(config map[string]interface{}, tier vaultmodels.HashicorpCloudVault20201125Tier) (*vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfig, diag.Diagnostics) {
	if tier == vaultmodels.HashicorpCloudVault20201125TierDEV || tier == vaultmodels.HashicorpCloudVault20201125TierSTARTERSMALL {
		return nil, diag.Errorf("major version configuration is only allowed for STANDARD or PLUS clusters")
	}

	mvuConfig := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfig{}

	upgradeType := config["upgrade_type"].(string)
	mvuConfigpgradeType := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigUpgradeType(upgradeType)
	mvuConfig.UpgradeType = &mvuConfigpgradeType

	maintenanceWindowDay := config["maintenance_window_day"].(string)
	maintenanceWindowTime := config["maintenance_window_time"].(string)

	if *mvuConfig.UpgradeType == vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigUpgradeTypeSCHEDULED {
		if maintenanceWindowDay == "" || maintenanceWindowTime == "" {
			return nil, diag.Errorf("major version upgrade configuration is invalid: maintenance window configuration information missing")
		}
		dayOfWeek := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindowDayOfWeek(maintenanceWindowDay)
		timeWindowUtc := vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindowTimeWindowUTC(maintenanceWindowTime)
		mvuConfig.MaintenanceWindow = &vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigMaintenanceWindow{
			DayOfWeek:     &dayOfWeek,
			TimeWindowUtc: &timeWindowUtc,
		}
	} else {
		if maintenanceWindowDay != "" || maintenanceWindowTime != "" {
			return nil, diag.Errorf("major version upgrade configuration is invalid: maintenance window is only allowed to SCHEDULED upgrades")
		}
		mvuConfig.MaintenanceWindow = nil
	}

	return &mvuConfig, nil
}

func flattenMajorVersionUpgradeConfig(config *vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfig, d *schema.ResourceData) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	configMap := map[string]interface{}{}
	upgradeType := config.UpgradeType

	configMap["upgrade_type"] = upgradeType
	if *upgradeType == vaultmodels.HashicorpCloudVault20201125MajorVersionUpgradeConfigUpgradeTypeSCHEDULED && config.MaintenanceWindow != nil {
		configMap["maintenance_window_day"] = config.MaintenanceWindow.DayOfWeek
		configMap["maintenance_window_time"] = config.MaintenanceWindow.TimeWindowUtc
	}

	return []interface{}{configMap}
}

func resourceVaultClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	clusterID := d.Id()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	link := newLink(loc, VaultClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}

func inPlusTier(tier string) bool {
	return tier == string(vaultmodels.HashicorpCloudVault20201125TierPLUSSMALL) ||
		tier == string(vaultmodels.HashicorpCloudVault20201125TierPLUSMEDIUM) ||
		tier == string(vaultmodels.HashicorpCloudVault20201125TierPLUSLARGE)
}

func validatePerformanceReplicationChecksAndReturnPrimaryIfAny(ctx context.Context, client *clients.Client, d *schema.ResourceData) (diag.Diagnostics, *vaultmodels.HashicorpCloudVault20201125Cluster) {
	primaryClusterLinkStr := getPrimaryLinkIfAny(d)
	// If no primary_link has been supplied, treat this as as single cluster creation.
	if primaryClusterLinkStr == "" {
		return nil, nil
	}

	primaryCluster, err := getPrimaryClusterFromLink(ctx, client, primaryClusterLinkStr)
	if err != nil {
		return err, nil
	}

	if !inPlusTier(string(*primaryCluster.Config.Tier)) {
		return diag.Errorf("primary cluster (%s) must be plus-tier", primaryCluster.ID), primaryCluster
	}

	// Tier should be specified, even if secondary inherits it from the primary cluster.
	if !strings.EqualFold(d.Get("tier").(string), string(*primaryCluster.Config.Tier)) {
		return diag.Errorf("a secondary's tier must match that of its primary (%s)", primaryCluster.ID), primaryCluster
	}

	if primaryCluster.PerformanceReplicationInfo != nil && *primaryCluster.PerformanceReplicationInfo.Mode == vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationInfoModeSECONDARY {
		return diag.Errorf("primary cluster (%s) is already a secondary", primaryCluster.ID), primaryCluster
	}

	// min_vault_version should either be empty or match the primary's initial version
	minVaultVersion := d.Get("min_vault_version").(string)
	if minVaultVersion != "" && !strings.EqualFold(minVaultVersion, primaryCluster.Config.VaultConfig.InitialVersion) {
		return diag.Errorf("min_vault_version should either be unset or match the primary cluster's (%s) initial version (%s)", primaryCluster.ID, primaryCluster.Config.VaultConfig.InitialVersion), primaryCluster
	}
	return nil, primaryCluster
}

func getPrimaryLinkIfAny(d *schema.ResourceData) string {
	primaryClusterLinkIface, ok := d.GetOk("primary_link")
	if !ok {
		return ""
	}
	return primaryClusterLinkIface.(string)
}

func getPrimaryClusterFromLink(ctx context.Context, client *clients.Client, link string) (*vaultmodels.HashicorpCloudVault20201125Cluster, diag.Diagnostics) {
	primaryClusterLink, err := buildLinkFromURL(link, VaultClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return nil, diag.Errorf("invalid primary_link supplied %v", err)
	}

	primaryCluster, err := clients.GetVaultClusterByID(ctx, client, primaryClusterLink.Location, primaryClusterLink.ID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			return nil, diag.Errorf("primary cluster (%s) does not exist", primaryClusterLink.ID)

		}
		return nil, diag.Errorf("unable to check for presence of an existing primary Vault cluster (%s): %v", primaryClusterLink.ID, err)
	}
	return primaryCluster, nil
}

func getPathStrings(pathFilter interface{}) []string {
	pathFilterArr := pathFilter.([]interface{})
	var paths []string
	for _, pathFilter := range pathFilterArr {
		paths = append(paths, pathFilter.(string))
	}
	return paths
}

func printPlusScalingWarningMsg() {
	log.Printf("[WARN] When scaling Plus-tier Vault clusters, be sure to keep the size of all clusters in a replication group in sync")
}
