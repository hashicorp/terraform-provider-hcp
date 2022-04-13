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
var createUpdateVaultClusterTimeout = time.Minute * 35

// deleteVaultClusterTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteVaultClusterTimeout = time.Minute * 25

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
				Description:      "Tier of the HCP Vault cluster. Valid options for tiers - `dev`, `starter_small`, `standard_small`, `standard_medium`, `standard_large`, `plus_small`, `plus_medium`, `plus_large`. See [pricing information](https://cloud.hashicorp.com/pricing/vault).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateVaultClusterTier,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
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
				Description: "The performance replication [paths filter](https://learn.hashicorp.com/tutorials/vault/paths-filter). Applies to performance replication secondaries only and operates in \"deny\" mode only.",
				Type:        schema.TypeList,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validateVaultPathsFilter,
				},
				Optional: true,
				Computed: true,
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
			"vault_version": {
				Description: "The Vault version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
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
		if paths, ok := d.GetOk("paths_filter"); ok {
			pathStrings := getPathStrings(paths)
			pathsFilter = &vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter{
				Mode:  vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilterModeDENY,
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

		vaultCluster = &vaultmodels.HashicorpCloudVault20201125InputCluster{
			Config: &vaultmodels.HashicorpCloudVault20201125InputClusterConfig{
				VaultConfig: &vaultmodels.HashicorpCloudVault20201125VaultConfig{
					InitialVersion: vaultVersion,
				},
				Tier: vaultmodels.HashicorpCloudVault20201125Tier(strings.ToUpper(d.Get("tier").(string))),
				NetworkConfig: &vaultmodels.HashicorpCloudVault20201125InputNetworkConfig{
					NetworkID:        hvn.ID,
					PublicIpsEnabled: publicEndpoint,
				},
			},
			ID:       clusterID,
			Location: loc,
		}
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

	// The Vault cluster failed to provision properly so we want to let the user know and
	// remove it from state.
	if cluster.State == vaultmodels.HashicorpCloudVault20201125ClusterStateFAILED {
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

	// Confirm at least on modifiable field has changed
	if !d.HasChanges("tier", "public_endpoint", "paths_filter") {
		return nil
	}

	if d.HasChange("tier") {
		clusterToScale := cluster
		destTier := vaultmodels.HashicorpCloudVault20201125Tier(strings.ToUpper(d.Get("tier").(string)))
		if inPlusTier(string(cluster.Config.Tier)) {
			// Plus tier clusters scale as a group via the primary cluster.
			// However, it is still worth individually tracking the tier of each cluster so that the
			// provider has the same information as the portal UI and can detect a scaling operation that
			// fails part way through to enable retries.
			// Because the clusters scale as group,
			//  a) replicated clusters may have already scaled due to another resource's update
			//  b) all scaling requests are routed through the primary
			// It is important to keep the tier of all replicated clusters in sync.

			// Because of (a), check that the scaling operation is necessary.
			if cluster.Config.Tier == destTier {
				clusterToScale = nil
			} else {
				printPlusScalingWarningMsg()
				primaryLink := getPrimaryLinkIfAny(d)
				if primaryLink != "" {
					// Because of (b), if the cluster is a secondary, issue the actual API request to the primary.
					var getPrimaryErr diag.Diagnostics
					clusterToScale, getPrimaryErr = getPrimaryClusterFromLink(ctx, client, primaryLink)
					if getPrimaryErr != nil {
						return getPrimaryErr
					}
				}
			}
		}

		if clusterToScale != nil {
			// Invoke update tier endpoint.
			updateResp, err := clients.UpdateVaultClusterTier(ctx, client, clusterToScale.Location, clusterToScale.ID, destTier)
			if err != nil {
				return diag.Errorf("error updating Vault cluster tier (%s): %v", clusterID, err)
			}

			// Wait for the update cluster operation.
			if err := clients.WaitForOperation(ctx, client, "update Vault cluster tier", clusterToScale.Location, updateResp.Operation.ID); err != nil {
				return diag.Errorf("unable to update Vault cluster tier (%s): %v", clusterID, err)
			}
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
			updateResp, err := clients.UpdateVaultPathsFilter(ctx, client, cluster.Location, clusterID, vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilter{
				Mode:  vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationPathsFilterModeDENY,
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

	publicEndpoint := cluster.Config.NetworkConfig.PublicIpsEnabled
	if err := d.Set("public_endpoint", publicEndpoint); err != nil {
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
			d.Set("paths_filter", nil)
		}
	}

	return nil
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

	if !inPlusTier(string(primaryCluster.Config.Tier)) {
		return diag.Errorf("primary cluster (%s) must be plus-tier", primaryCluster.ID), primaryCluster
	}

	// Tier should be specified, even if secondary inherits it from the primary cluster.
	if !strings.EqualFold(d.Get("tier").(string), string(primaryCluster.Config.Tier)) {
		return diag.Errorf("a secondary's tier must match that of its primary (%s)", primaryCluster.ID), primaryCluster
	}

	if primaryCluster.PerformanceReplicationInfo != nil && primaryCluster.PerformanceReplicationInfo.Mode == vaultmodels.HashicorpCloudVault20201125ClusterPerformanceReplicationInfoModeSECONDARY {
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
