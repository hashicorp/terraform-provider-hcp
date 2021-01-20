package provider

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/client/consul_service"
	consulmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-consul-service/preview/2020-08-26/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/consul"
	"github.com/hashicorp/terraform-provider-hcp/internal/helper"
)

// defaultClusterTimeoutDuration is the amount of time that can elapse
// before a cluster read operation should timeout.
var defaultClusterTimeoutDuration = time.Minute * 5

// createUpdateTimeoutDuration is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateTimeoutDuration = time.Minute * 30

// deleteTimeoutDuration is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteTimeoutDuration = time.Minute * 25

// consulCusterResourceCloudProviders is the list of cloud providers
// where a HCP Consul cluster can be provisioned.
var consulCusterResourceCloudProviders = []string{
	"aws",
}

// resourceConsulCluster represents an HCP Consul cluster.
func resourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The Consul cluster resource allow you to manage an HCP Consul cluster.",
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
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"hvn_id": {
				Description:      "The ID of the HVN this HCP Consul cluster is associated to.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
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
			// optional fields
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint for the Consul UI. Defaults to false.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"project_id": {
				Description: "The ID of the project this HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
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
			"consul_connect": {
				Description: "Denotes that Consul connect is enabled.",
				Type:        schema.TypeBool,
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
		},
	}
}

func resourceConsulClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client)

	clusterID := d.Get("cluster_id").(string)

	loc, err := helper.BuildResourceLocation(ctx, d, client, "Consul cluster")
	if err != nil {
		return diag.FromErr(err)
	}

	// Check for an existing Consul cluster
	_, err = clients.GetConsulClusterByID(ctx, client, loc, clusterID)
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
	availableConsulVersions, err := clients.GetAvailableHCPConsulVersions(ctx, client)
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

	// explicitly set min_consul_version so it will be populated in resource data if not specified as an input
	if err := d.Set("min_consul_version", consulVersion); err != nil {
		return diag.Errorf("error determining min_consul_version: %v", err)
	}

	datacenter := clusterID
	v, ok = d.GetOk("datacenter")
	if ok {
		datacenter = v.(string)
	}

	connectEnabled := d.Get("connect_enabled").(bool)
	publicEndpoint := d.Get("public_endpoint").(bool)

	// TODO more explicit logic once tier levels are fleshed out
	numServers := int32(1)

	hvnID := d.Get("hvn_id").(string)

	createConsulClusterParams := consul_service.NewCreateParams()
	createConsulClusterParams.Context = ctx
	createConsulClusterParams.Body = &consulmodels.HashicorpCloudConsul20200826CreateRequest{
		Cluster: &consulmodels.HashicorpCloudConsul20200826Cluster{
			Config: &consulmodels.HashicorpCloudConsul20200826ClusterConfig{
				CapacityConfig: &consulmodels.HashicorpCloudConsul20200826CapacityConfig{
					NumServers: numServers,
				},
				ConsulConfig: &consulmodels.HashicorpCloudConsul20200826ConsulConfig{
					ConnectEnabled: connectEnabled,
					Datacenter:     datacenter,
				},
				MaintenanceConfig: nil,
				NetworkConfig: &consulmodels.HashicorpCloudConsul20200826NetworkConfig{
					Network: newLink(loc, "hvn", hvnID),
					Private: !publicEndpoint,
				},
			},
			ConsulVersion: consulVersion,
			ID:            clusterID,
			Location:      loc,
		},
	}

	createConsulClusterParams.ClusterLocationOrganizationID = loc.OrganizationID
	createConsulClusterParams.ClusterLocationProjectID = loc.ProjectID

	log.Printf("[INFO] Creating Consul cluster (%s)", clusterID)

	createClusterResp, err := client.Consul.Create(createConsulClusterParams, nil)
	if err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", clusterID, err)
	}

	// wait for the Consul cluster to be created
	if err := clients.WaitForOperation(ctx, client, "create Consul cluster", loc, createClusterResp.Payload.Operation.ID); err != nil {
		return diag.Errorf("unable to create Consul cluster (%s): %v", createClusterResp.Payload.Cluster.ID, err)
	}

	log.Printf("[INFO] Created Consul cluster (%s)", createClusterResp.Payload.Cluster.ID)

	// get the created Consul cluster
	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, createClusterResp.Payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster (%s): %v", createClusterResp.Payload.Cluster.ID, err)
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, createClusterResp.Payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to retrieve Consul cluster client config files (%s): %v", createClusterResp.Payload.Cluster.ID, err)
	}

	// create customer master ACL token
	masterACLToken, err := clients.CreateCustomerMasterACLToken(ctx, client, loc, createClusterResp.Payload.Cluster.ID)
	if err != nil {
		return diag.Errorf("unable to create master ACL token for cluster (%s): %v", createClusterResp.Payload.Cluster.ID, err)
	}

	// Only set root token keys after create
	if err := d.Set("consul_root_token_accessor_id", masterACLToken.ACLToken.AccessorID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("consul_root_token_secret_id", masterACLToken.ACLToken.SecretID); err != nil {
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
	link := newLink(cluster.Location, "consul-service", cluster.ID)
	url, err := linkURL(link)
	if err != nil {
		return err
	}

	d.SetId(url)

	if err := d.Set("hvn_id", cluster.Config.NetworkConfig.Network.ID); err != nil {
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

	if err := d.Set("project_id", cluster.Location.ProjectID); err != nil {
		return err
	}

	if err := d.Set("datacenter", cluster.Config.ConsulConfig.Datacenter); err != nil {
		return err
	}

	if err := d.Set("consul_automatic_upgrades", true); err != nil {
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

	if err := d.Set("consul_connect", cluster.Config.ConsulConfig.ConnectEnabled); err != nil {
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
	return nil
}

func resourceConsulClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceConsulClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
