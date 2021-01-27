package provider

import (
	"context"
	"log"
	"strings"
	"time"

	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"

	consulmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/consul"
	"github.com/hashicorp/terraform-provider-hcp/internal/helper"
)

const consulTierDevelopment = "development"
const consulTierStandard = "standard"

// defaultClusterTimeoutDuration is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultClusterTimeoutDuration = time.Minute * 5

// createUpdateTimeoutDuration is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateTimeoutDuration = time.Minute * 35

// deleteTimeoutDuration is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteTimeoutDuration = time.Minute * 25

// consulCusterResourceCloudProviders is the list of cloud providers
// where a HCP Consul cluster can be provisioned.
var consulCusterResourceCloudProviders = []string{
	"aws",
}

// consulClusterResourceTiers is the list of tiers
// that an HCP Consul cluster can be provisioned as.
var consulClusterResourceTiers = []string{
	consulTierDevelopment,
	consulTierStandard,
}

// tierToNumServers maps the set of tiers
// to the number of servers to be provisioned for that cluster.
var consulTierToNumServers = map[string]int32{
	consulTierDevelopment: int32(1),
	consulTierStandard:    int32(3),
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
			Default: &defaultClusterTimeoutDuration,
			Create:  &createUpdateTimeoutDuration,
			Update:  &createUpdateTimeoutDuration,
			Delete:  &deleteTimeoutDuration,
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
			"cloud_provider": {
				Description:      "The provider where the HCP Consul cluster is located. Only 'aws' is available at this time.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSlice(consulCusterResourceCloudProviders, true),
			},
			"region": {
				Description:      "The region where the HCP Consul cluster is located.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"tier": {
				// TODO: link to HCP Consul feature tier page when it is available
				Description:      "The tier that the HCP Consul cluster will be provisioned as.  Only 'development' and 'standard' are available at this time.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringInSlice(consulClusterResourceTiers, true),
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
					return consul.NormalizeVersion(old) == consul.NormalizeVersion(new) || new == ""
				},
			},
			"datacenter": {
				Description: "The Consul data center name of the cluster. If not specified, it is defaulted to the value of `cluster_id`.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"connect_enabled": {
				Description: "Denotes the Consul connect feature should be enabled for this cluster.  Default to true.",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				ForceNew:    true,
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
		},
	}
}

func resourceConsulClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
		Region: &sharedmodels.HashicorpCloudLocationRegion{
			Provider: d.Get("cloud_provider").(string),
			Region:   d.Get("region").(string),
		},
	}

	// Check for an existing Consul cluster
	_, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		if !clients.IsResponseCodeNotFound(err) {
			return diag.Errorf("unable to check for presence of an existing Consul Cluster (%s): %v", clusterID, err)
		}

		// a 404 indicates a Consul cluster was not found
		log.Printf("[INFO] Consul cluster (%s) not found, proceeding with create", clusterID)
	} else {
		return diag.Errorf("a Consul cluster with cluster_id=%q in project_id=%q already exists - to be managed via Terraform this resource needs to be imported into the State.  Please see the resource documentation for hcp_consul_cluster for more information.", clusterID, loc.ProjectID)
	}

	// fetch available version from HCP
	availableConsulVersions, err := clients.GetAvailableHCPConsulVersions(ctx, loc, client)
	if err != nil || availableConsulVersions == nil {
		return diag.Errorf("error fetching available HCP Consul versions: %v", err)
	}

	// determine recommended version
	consulVersion := consul.RecommendedVersion(availableConsulVersions)
	v, ok := d.GetOk("min_consul_version")
	if ok {
		consulVersion = consul.NormalizeVersion(v.(string))
	}

	// check if version is valid and available
	if !consul.IsValidVersion(consulVersion, availableConsulVersions) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %v", consulVersion, availableConsulVersions)
	}

	datacenter := clusterID
	v, ok = d.GetOk("datacenter")
	if ok {
		datacenter = v.(string)
	}

	connectEnabled := d.Get("connect_enabled").(bool)
	publicEndpoint := d.Get("public_endpoint").(bool)

	tier := d.Get("tier").(string)
	numServers := consulTierToNumServers[tier]

	hvnID := d.Get("hvn_id").(string)

	log.Printf("[INFO] Creating Consul cluster (%s)", clusterID)

	payload, err := clients.CreateConsulCluster(ctx, client, loc, clusterID, datacenter, consulVersion,
		numServers, !publicEndpoint, connectEnabled, newLink(loc, "hvn", hvnID))
	if err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", clusterID, err)
	}

	// wait for the Consul cluster to be created
	if err := clients.WaitForOperation(ctx, client, "create Consul cluster", loc, payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", payload.Cluster.ID, err)
	}

	log.Printf("[INFO] Created Consul cluster (%s)", payload.Cluster.ID)

	link := newLink(loc, ConsulClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// get the created Consul cluster
	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s): %v", payload.Cluster.ID, err)
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster client config files (%s): %v", payload.Cluster.ID, err)
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
func setConsulClusterResourceData(d *schema.ResourceData, cluster *consulmodels.HashicorpCloudConsul20200826Cluster,
	clientConfigFiles *consulmodels.HashicorpCloudConsul20200826GetClientConfigResponse) error {

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

	if err := d.Set("scale", cluster.Config.CapacityConfig.NumServers); err != nil {
		return err
	}

	// TODO: Update this logic when tier becomes a first class value on the cluster
	for t, numServers := range consulTierToNumServers {
		if numServers == cluster.Config.CapacityConfig.NumServers {
			if err := d.Set("tier", t); err != nil {
				return err
			}
			break
		}
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
		if err := d.Set("consul_public_endpoint_url", cluster.DNSNames.Public); err != nil {
			return err
		}
	}

	if err := d.Set("consul_private_endpoint_url", cluster.DNSNames.Private); err != nil {
		return err
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

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, clusterID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster client config files (%s): %v", clusterID, err)
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
	newConsulVersion := consul.NormalizeVersion(v.(string))

	// Check that there are any valid upgrade versions
	if upgradeVersions == nil {
		return diag.Errorf("no upgrade versions of Consul are available for this cluster; you may already be on the latest Consul version supported by HCP")
	}

	// Validate that the upgrade version is valid
	if !consul.IsValidVersion(newConsulVersion, upgradeVersions) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %v", newConsulVersion, upgradeVersions)
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
		return diag.Errorf("unable to retrieve Consul cluster client config files (%s): %v", clusterID, err)
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
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	link := newLink(loc, ConsulClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return nil, err
	}
	d.SetId(url)

	diags := resourceConsulClusterRead(ctx, d, meta)
	if err := helper.ToError(diags); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
