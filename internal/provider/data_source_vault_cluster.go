// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func dataSourceVaultCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster data source provides information about an existing HCP Vault cluster.",
		ReadContext: dataSourceVaultClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultVaultClusterTimeout,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"cluster_id": {
				Description:      "The ID of the HCP Vault cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// computed outputs
			"hvn_id": {
				Description: "The ID of the HVN this HCP Vault cluster is associated to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public_endpoint": {
				Description: "Denotes that the cluster has a public endpoint. Defaults to false.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"min_vault_version": {
				Description: "The minimum Vault version to use when creating the cluster. If not specified, it is defaulted to the version that is currently recommended by HCP.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tier": {
				Description: "The tier that the HCP Vault cluster will be provisioned as.  Only 'development' is available at this time.",
				Type:        schema.TypeString,
				Computed:    true,
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
			"state": {
				Description: "The state of the Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_link": {
				Description: "The `self_link` of the HCP Vault Plus tier cluster which is the primary in the performance replication setup with this HCP Vault Plus tier cluster. If not specified, it is a standalone Plus tier HCP Vault cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"paths_filter": {
				Description: "The performance replication [paths filter](https://developer.hashicorp.com/vault/tutorials/cloud-ops/vault-replication-terraform#review-hcpvault-tf). Applies to performance replication secondaries only and operates in \"deny\" mode only.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"metrics_config": {
				Description: "The metrics configuration for export. (https://developer.hashicorp.com/vault/tutorials/cloud-monitoring/vault-metrics-guide#metrics-streaming-configuration)",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grafana_endpoint": {
							Description: "Grafana endpoint for streaming metrics",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"grafana_user": {
							Description: "Grafana user for streaming metrics",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"splunk_hecendpoint": {
							Description: "Splunk endpoint for streaming metrics",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"datadog_region": {
							Description: "Datadog region for streaming metrics",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"audit_log_config": {
				Description: "The audit logs configuration for export. (https://developer.hashicorp.com/vault/tutorials/cloud-monitoring/vault-metrics-guide#metrics-streaming-configuration)",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grafana_endpoint": {
							Description: "Grafana endpoint for streaming audit logs",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"grafana_user": {
							Description: "Grafana user for streaming audit logs",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"splunk_hecendpoint": {
							Description: "Splunk endpoint for streaming audit logs",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"datadog_region": {
							Description: "Datadog region for streaming audit logs",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"major_version_upgrade_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"upgrade_type": {
							Description: "The major upgrade type for the cluster",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"maintenance_window_day": {
							Description: "The maintenance day of the week for scheduled updates",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"maintenance_window_time": {
							Description: "The maintenance time frame for scheduled updates",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVaultClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterID := d.Get("cluster_id").(string)
	client := meta.(*clients.Client)

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	log.Printf("[INFO] Reading Vault cluster (%s) [project_id=%s, organization_id=%s]", clusterID, loc.ProjectID, loc.OrganizationID)

	cluster, err := clients.GetVaultClusterByID(ctx, client, loc, clusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	// build the id for this Vault cluster
	link := newLink(loc, VaultClusterResourceType, clusterID)
	url, err := linkURL(link)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(url)

	// Cluster found, update resource data.
	if err := setVaultClusterResourceData(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
