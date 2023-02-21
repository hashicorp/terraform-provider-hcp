// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"

	consulmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/stable/2021-02-04/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceConsulCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster data source provides information about an existing HCP Consul cluster.",
		ReadContext: dataSourceConsulClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultConsulClusterTimeout,
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
			"state": {
				Description: "The state of the HCP Consul cluster.",
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
				Description: "The tier that the HCP Consul cluster will be provisioned as.  Only `development`, `standard` and `plus` are available at this time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The t-shirt size representation of each server VM that this Consul cluster is provisioned with. Valid option for development tier - `x_small`. Valid options for other tiers - `small`, `medium`, `large`. For more details - https://cloud.hashicorp.com/pricing/consul",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"self_link": {
				Description: "A unique URL identifying the HCP Consul cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_link": {
				Description: "The `self_link` of the HCP Consul cluster which is the primary in the federation setup with this HCP Consul cluster. If not specified, it is a standalone cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auto_hvn_to_hvn_peering": {
				Description: "Enables automatic HVN to HVN peering when creating a secondary cluster in a federation.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ip_allowlist": {
				Description: "Allowed IPV4 address ranges (CIDRs) for inbound traffic. Each entry must be a unique CIDR. Maximum 3 CIDRS supported at this time.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Description: "IP address range in CIDR notation.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "Description to help identify source (maximum 255 chars).",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
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

	// cluster found, set data source attributes
	if err := setConsulClusterDataSourceAttributes(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	// get the cluster's Consul client config files
	clientConfigFiles, err := clients.GetConsulClientConfigFiles(ctx, client, loc, clusterID)
	if err != nil {
		log.Printf("[WARN] unable to retrieve Consul cluster (%s) client config files: %v", clusterID, err)
		return nil
	}

	// client config found, set data source attributes
	if err := setConsulClusterClientConfigDataSourceAttributes(d, clientConfigFiles); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// setConsulClusterDataSourceAttributes sets all data source attributes from the cluster
func setConsulClusterDataSourceAttributes(
	d *schema.ResourceData,
	cluster *consulmodels.HashicorpCloudConsul20210204Cluster,
) error {

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

	if err := d.Set("state", cluster.State); err != nil {
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

	if err := d.Set("connect_enabled", cluster.Config.ConsulConfig.ConnectEnabled); err != nil {
		return err
	}

	if err := d.Set("consul_version", cluster.ConsulVersion); err != nil {
		return err
	}

	if err := d.Set("auto_hvn_to_hvn_peering", cluster.Config.AutoHvnToHvnPeering); err != nil {
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
	selfLink, err := linkURL(link)
	if err != nil {
		return err
	}
	if err := d.Set("self_link", selfLink); err != nil {
		return err
	}

	if cluster.Config.ConsulConfig.Primary != nil {
		link := newLink(cluster.Config.ConsulConfig.Primary.Location, ConsulClusterResourceType, cluster.Config.ConsulConfig.Primary.ID)
		primaryLink, err := linkURL(link)
		if err != nil {
			return err
		}
		if err := d.Set("primary_link", primaryLink); err != nil {
			return err
		}
	}

	if cluster.Config.NetworkConfig != nil {
		ipAllowlist := make([]map[string]interface{}, len(cluster.Config.NetworkConfig.IPAllowlist))
		for i, cidrRange := range cluster.Config.NetworkConfig.IPAllowlist {
			cidr := map[string]interface{}{
				"description": cidrRange.Description,
				"address":     cidrRange.Address,
			}
			ipAllowlist[i] = cidr
		}
		if err := d.Set("ip_allowlist", ipAllowlist); err != nil {
			return err
		}
	}

	return nil
}

// setConsulClusterClientConfigDataSourceAttributes sets all resource data that's derived from client config meta
func setConsulClusterClientConfigDataSourceAttributes(
	d *schema.ResourceData,
	clientConfigFiles *consulmodels.HashicorpCloudConsul20210204GetClientConfigResponse,
) error {
	if err := d.Set("consul_config_file", clientConfigFiles.ConsulConfigFile.String()); err != nil {
		return err
	}

	if err := d.Set("consul_ca_file", clientConfigFiles.CaFile.String()); err != nil {
		return err
	}

	return nil
}
