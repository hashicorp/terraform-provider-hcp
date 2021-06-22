package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2021-02-04/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/consul"
	"github.com/hashicorp/terraform-provider-hcp/internal/input"
)

// defaultClusterTimeout is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultConsulClusterTimeout = time.Minute * 5

// createUpdateTimeout is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateConsulClusterTimeout = time.Minute * 35

// deleteTimeout is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteConsulClusterTimeout = time.Minute * 25

// consulCusterResourceCloudProviders is the list of cloud providers
// where a HCP Consul cluster can be provisioned.
var consulCusterResourceCloudProviders = []string{
	"aws",
}

// resourceConsulCluster represents an HCP Consul cluster.
func resourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The Consul cluster resource allows you to manage an HCP Consul cluster.",
		CreateContext: resourceConsulClusterCreate,
		ReadContext:   resourceConsulClusterRead,
		UpdateContext: resourceConsulClusterUpdate,
		DeleteContext: resourceConsulClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultConsulClusterTimeout,
			Create:  &createUpdateConsulClusterTimeout,
			Update:  &createUpdateConsulClusterTimeout,
			Delete:  &deleteConsulClusterTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceConsulClusterImport,
		},
		Schema: map[string]*schema.Schema{
			// required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"hvn_id": {
				Description:      "The ID of the HVN this HCP Consul cluster is associated to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"tier": {
				Description:      "The tier that the HCP Consul cluster will be provisioned as.  Only `development`, `standard` and `plus` are available at this time. See [pricing information](https://cloud.hashicorp.com/pricing/consul).",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateConsulClusterTier,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			// optional fields
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint for the Consul UI. Defaults to false.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"min_consul_version": {
				Description:      "The minimum Consul version of the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSemVer,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// Suppress diff is normalized versions match OR min_consul_version is removed from the resource
					// since min_consul_version is required in order to upgrade the cluster to a new Consul version.
					return input.NormalizeVersion(old) == input.NormalizeVersion(new) || new == ""
				},
			},
			"datacenter": {
				Description:      "The Consul data center name of the cluster. If not specified, it is defaulted to the value of `cluster_id`.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateDatacenter,
			},
			"connect_enabled": {
				Description: "Denotes the Consul connect feature should be enabled for this cluster.  Default to true.",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				ForceNew:    true,
			},
			"primary_link": {
				Description: "The `self_link` of the HCP Consul cluster which is the primary in the federation setup with this HCP Consul cluster. If not specified, it is a standalone cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"size": {
				Description:      "The t-shirt size representation of each server VM that this Consul cluster is provisioned with. Valid option for development tier - `x_small`. Valid options for other tiers - `small`, `medium`, `large`. For more details - https://cloud.hashicorp.com/pricing/consul",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateConsulClusterSize,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			// computed outputs
			"organization_id": {
				Description: "The ID of the organization this HCP Consul cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Consul cluster is located in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_provider": {
				Description: "The provider where the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "The region where the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_automatic_upgrades": {
				Description: "Denotes that automatic Consul upgrades are enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"consul_snapshot_interval": {
				Description: "The Consul snapshot interval.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_snapshot_retention": {
				Description: "The retention policy for Consul snapshots.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_config_file": {
				Description: "The cluster config encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_ca_file": {
				Description: "The cluster CA file encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_version": {
				Description: "The Consul version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_public_endpoint_url": {
				Description: "The public URL for the Consul UI. This will be empty if `public_endpoint` is `false`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_private_endpoint_url": {
				Description: "The private URL for the Consul UI.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_root_token_accessor_id": {
				Description: "The accessor ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_root_token_secret_id": {
				Description: "The secret ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcp_consul_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"scale": {
				Description: "The number of Consul server nodes in the cluster.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the HCP Consul cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConsulClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Check for an existing Consul cluster
	_, err = clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Consul cluster (%s): %v", clusterID, err)
		}

		// a 404 indicates a Consul cluster was not found
		log.Printf("[INFO] Consul cluster (%s) not found, proceeding with create", clusterID)
	} else {
		return diag.Errorf("a Consul cluster with cluster_id=%q in project_id=%q already exists - to be managed via Terraform this resource needs to be imported into the State.  Please see the resource documentation for hcp_consul_cluster for more information.", clusterID, loc.ProjectID)
	}

	// fetch available version from HCP
	availableConsulVersions, err := clients.GetAvailableHCPConsulVersionsForLocation(ctx, loc, client)
	if err != nil || availableConsulVersions == nil {
		return diag.Errorf("error fetching available HCP Consul versions: %v", err)
	}

	// determine recommended version
	consulVersion := consul.RecommendedVersion(availableConsulVersions)
	v, ok := d.GetOk("min_consul_version")
	if ok {
		consulVersion = input.NormalizeVersion(v.(string))
	}

	// check if version is valid and available
	if !consul.IsValidVersion(consulVersion, availableConsulVersions) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: [%s]", consulVersion, consul.VersionsToString(availableConsulVersions))
	}

	// If specified, validate and parse the primary link provided for federation.
	primary_link, ok := d.GetOk("primary_link")
	var primary *sharedmodels.HashicorpCloudLocationLink
	if ok {
		primary, err = parseLinkURL(primary_link.(string), ConsulClusterResourceType)
		if err != nil {
			return diag.Errorf(err.Error())
		}
		primaryOrgID, err := clients.GetParentOrganizationIDByProjectID(ctx, client, primary.Location.ProjectID)
		if err != nil {
			return diag.Errorf("Error determining organization of primary cluster. %v", err)
		}
		primary.Location.OrganizationID = primaryOrgID
		// fetch the primary cluster
		primaryConsulCluster, err := clients.GetConsulClusterByID(ctx, client, primary.Location, primary.ID)
		if err != nil {
			return diag.Errorf("unable to check for presence of an existing primary Consul cluster (%s): %v", primary.ID, err)
		}
		primary.Location.Region = primaryConsulCluster.Location.Region
	}

	datacenter := strings.ToLower(clusterID)
	v, ok = d.GetOk("datacenter")
	if ok {
		datacenter = v.(string)
	}

	connectEnabled := d.Get("connect_enabled").(bool)
	publicEndpoint := d.Get("public_endpoint").(bool)

	log.Printf("[INFO] Creating Consul cluster (%s)", clusterID)

	consulCuster := &consulmodels.HashicorpCloudConsul20210204Cluster{
		Config: &consulmodels.HashicorpCloudConsul20210204ClusterConfig{
			Tier: consulmodels.HashicorpCloudConsul20210204ClusterConfigTier(strings.ToUpper(d.Get("tier").(string))),
			CapacityConfig: &consulmodels.HashicorpCloudConsul20210204CapacityConfig{
				Size: consulmodels.HashicorpCloudConsul20210204CapacityConfigSize(strings.ToUpper(d.Get("size").(string))),
			},
			ConsulConfig: &consulmodels.HashicorpCloudConsul20210204ConsulConfig{
				ConnectEnabled: connectEnabled,
				Datacenter:     datacenter,
				Primary:        primary,
			},
			MaintenanceConfig: nil,
			NetworkConfig: &consulmodels.HashicorpCloudConsul20210204NetworkConfig{
				Network: newLink(loc, "hvn", hvnID),
				Private: !publicEndpoint,
			},
		},
		ConsulVersion: consulVersion,
		ID:            clusterID,
		Location:      loc,
	}

	payload, err := clients.CreateConsulCluster(ctx, client, loc, consulCuster)
	if err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", clusterID, err)
	}

	link := newLink(loc, ConsulClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// wait for the Consul cluster to be created
	if err := clients.WaitForOperation(ctx, client, "create Consul cluster", loc, payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", payload.Cluster.ID, err)
	}

	log.Printf("[INFO] Created Consul cluster (%s)", payload.Cluster.ID)

	// get the created Consul cluster
	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s): %v", payload.Cluster.ID, err)
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s) client config files: %v", payload.Cluster.ID, err)
	}

	// create customer root ACL token
	rootACLToken, err := clients.CreateCustomerRootACLToken(ctx, client, loc, payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to create root ACL token for cluster (%s): %v", payload.Cluster.ID, err)
	}

	// Only set root token keys after create
	if err := d.Set("consul_root_token_accessor_id", rootACLToken.ACLToken.AccessorID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("consul_root_token_secret_id", rootACLToken.ACLToken.SecretID); err != nil {
		return diag.FromErr(err)
	}

	if err := setConsulClusterResourceData(d, cluster, clientConfigFiles); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// setConsulClusterResourceData sets the KV pairs of the Consul cluster resource schema.
// We do not set consul_root_token_accessor_id and consul_root_token_secret_id here since
// the original root token is only available during cluster creation.
func setConsulClusterResourceData(d *schema.ResourceData, cluster *consulmodels.HashicorpCloudConsul20210204Cluster,
	clientConfigFiles *consulmodels.HashicorpCloudConsul20210204GetClientConfigResponse) error {

	if err := d.Set("cluster_id", cluster.ID); err != nil {
		return err
	}

	if err := d.Set("hvn_id", cluster.Config.NetworkConfig.Network.ID); err != nil {
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

	publicEndpoint := !cluster.Config.NetworkConfig.Private
	if err := d.Set("public_endpoint", publicEndpoint); err != nil {
		return err
	}

	if err := d.Set("datacenter", cluster.Config.ConsulConfig.Datacenter); err != nil {
		return err
	}

	if err := d.Set("scale", cluster.Config.CapacityConfig.Scale); err != nil {
		return err
	}

	if err := d.Set("tier", cluster.Config.Tier); err != nil {
		return err
	}

	if err := d.Set("size", cluster.Config.CapacityConfig.Size); err != nil {
		return err
	}

	if err := d.Set("consul_snapshot_interval", "24h"); err != nil {
		return err
	}

	if err := d.Set("consul_snapshot_retention", "30d"); err != nil {
		return err
	}

	if err := d.Set("consul_config_file", clientConfigFiles.ConsulConfigFile.String()); err != nil {
		return err
	}

	if err := d.Set("consul_ca_file", clientConfigFiles.CaFile.String()); err != nil {
		return err
	}

	if err := d.Set("connect_enabled", cluster.Config.ConsulConfig.ConnectEnabled); err != nil {
		return err
	}

	if err := d.Set("consul_version", cluster.ConsulVersion); err != nil {
		return err
	}

	if publicEndpoint {
		// No port needed to communicate with HCP Consul via HTTPS
		if err := d.Set("consul_public_endpoint_url", fmt.Sprintf("https://%s", cluster.DNSNames.Public)); err != nil {
			return err
		}
	}

	// No port needed to communicate with HCP Consul via HTTPS
	if err := d.Set("consul_private_endpoint_url", fmt.Sprintf("https://%s", cluster.DNSNames.Private)); err != nil {
		return err
	}

	link := newLink(cluster.Location, ConsulClusterResourceType, cluster.ID)
	self_link, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", self_link); err != nil {
		return err
	}

	if cluster.Config.ConsulConfig.Primary != nil {
		link := newLink(cluster.Config.ConsulConfig.Primary.Location, ConsulClusterResourceType, cluster.Config.ConsulConfig.Primary.ID)
		primary_link, err := linkURL(link)
		if err != nil {
			return err
		}
		if err := d.Set("primary_link", primary_link); err != nil {
			return err
		}
	}

	return nil
}

func resourceConsulClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), ConsulClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Consul cluster (%s): %v", clusterID, err)
	}

	// The Consul cluster failed to provision properly so we want to let the user know and
	// remove it from state
	if cluster.State == consulmodels.HashicorpCloudConsul20210204ClusterStateFAILED {
		log.Printf("[WARN] Consul cluster (%s) failed to provision, removing from state", clusterID)
		d.SetId("")
		return nil
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s) client config files: %v", clusterID, err)
	}

	// Cluster found, update resource data
	if err := setConsulClusterResourceData(d, cluster, clientConfigFiles); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConsulClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), ConsulClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul cluster (%s) not found, removing from state", clusterID)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch Consul cluster (%s): %v", clusterID, err)
	}

	// Fetch available upgrade versions
	upgradeVersions, err := clients.ListConsulUpgradeVersions(ctx, client, cluster.Location, clusterID)
	if err != nil {
		return diag.Errorf("unable to list Consul upgrade versions (%s): %v", clusterID, err)
	}

	v, ok := d.GetOk("min_consul_version")
	if !ok {
		return diag.Errorf("min_consul_version is required in order to upgrade the cluster")
	}
	newConsulVersion := input.NormalizeVersion(v.(string))

	// Check that there are any valid upgrade versions
	if upgradeVersions == nil {
		return diag.Errorf("no upgrade versions of Consul are available for this cluster; you may already be on the latest Consul version supported by HCP")
	}

	// Validate that the upgrade version is valid
	if !consul.IsValidVersion(newConsulVersion, upgradeVersions) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: [%s]", newConsulVersion, consul.VersionsToString(upgradeVersions))
	}

	// Invoke update cluster endpoint
	updateResp, err := clients.UpdateConsulCluster(ctx, client, cluster.Location, clusterID, newConsulVersion)
	if err != nil {
		return diag.Errorf("error updating Consul cluster (%s): %v", clusterID, err)
	}

	// Wait for the update cluster operation
	if err := clients.WaitForOperation(ctx, client, "update Consul cluster", cluster.Location, updateResp.Operation.ID); err != nil {
		return diag.Errorf("unable to update Consul cluster (%s): %v", clusterID, err)
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, cluster.Location, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s) client config files: %v", clusterID, err)
	}

	if err := setConsulClusterResourceData(d, cluster, clientConfigFiles); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConsulClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	link, err := buildLinkFromURL(d.Id(), ConsulClusterResourceType, client.Config.OrganizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := link.ID
	loc := link.Location

	log.Printf("[INFO] Deleting Consul cluster (%s)", clusterID)

	deleteResp, err := clients.DeleteConsulCluster(ctx, client, loc, clusterID)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			log.Printf("[WARN] Consul cluster (%s) not found, so no action was taken", clusterID)
			return nil
		}

		return diag.Errorf("unable to delete Consul cluster (%s): %v", clusterID, err)
	}

	// Wait for the delete cluster operation
	if err := clients.WaitForOperation(ctx, client, "delete Consul cluster", loc, deleteResp.Operation.ID); err != nil {
		return diag.Errorf("unable to delete Consul cluster (%s): %v", clusterID, err)
	}

	return nil
}

func resourceConsulClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*clients.Client)

	clusterID := d.Id()
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		ProjectID: client.Config.ProjectID,
	}

	link := newLink(loc, ConsulClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}

	d.SetId(url)

	return []*schema.ResourceData{d}, nil
}
