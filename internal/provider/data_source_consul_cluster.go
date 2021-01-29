package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster data source provides information about an existing HCP Consul cluster",
		ReadContext: dataSourceConsulClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Consul cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// computed outputs
			"project_id": {
				Description: "The ID of the project this HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"organization_id": {
				Description: "The ID of the organization the project for this HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hvn_id": {
				Description: "The ID of the HVN this HCP Consul cluster is associated to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_provider": {
				Description: "The provider where the HCP Consul cluster is located. Only 'aws' is available at this time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "The region where the HCP Consul cluster is located.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint for the Consul UI. Defaults to false.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"datacenter": {
				Description: "The Consul data center name of the cluster. If not specified, it is defaulted to the value of `cluster_id`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connect_enabled": {
				Description: "Denotes the Consul connect feature should be enabled for this cluster.  Default to true.",
				Type:        schema.TypeBool,
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
			"scale": {
				Description: "The the number of Consul server nodes in the cluster.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"tier": {
				Description: "The tier that the HCP Consul cluster will be provisioned as.  Only 'development' and 'standard' are available at this time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceConsulClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterID := d.Get("cluster_id").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	log.Printf("[INFO] Reading Consul cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetConsulClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	// build the id for this Consul cluster
	link := newLink(loc, ConsulClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

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
