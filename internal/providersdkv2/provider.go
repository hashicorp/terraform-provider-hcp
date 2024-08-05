// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providersdkv2

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"hcp_aws_network_peering":            dataSourceAwsNetworkPeering(),
				"hcp_aws_transit_gateway_attachment": dataSourceAwsTransitGatewayAttachment(),
				"hcp_azure_peering_connection":       dataSourceAzurePeeringConnection(),
				"hcp_boundary_cluster":               dataSourceBoundaryCluster(),
				"hcp_consul_agent_helm_config":       dataSourceConsulAgentHelmConfig(),
				"hcp_consul_agent_kubernetes_secret": dataSourceConsulAgentKubernetesSecret(),
				"hcp_consul_cluster":                 dataSourceConsulCluster(),
				"hcp_consul_versions":                dataSourceConsulVersions(),
				"hcp_hvn":                            dataSourceHvn(),
				"hcp_hvn_peering_connection":         dataSourceHvnPeeringConnection(),
				"hcp_hvn_route":                      dataSourceHVNRoute(),
				"hcp_packer_bucket_names":            dataSourcePackerBucketNames(),
				"hcp_packer_run_task":                dataSourcePackerRunTask(),
				"hcp_vault_cluster":                  dataSourceVaultCluster(),
				"hcp_vault_plugin":                   dataSourceVaultPlugin(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hcp_aws_network_peering":            resourceAwsNetworkPeering(),
				"hcp_aws_transit_gateway_attachment": resourceAwsTransitGatewayAttachment(),
				"hcp_azure_peering_connection":       resourceAzurePeeringConnection(),
				"hcp_boundary_cluster":               resourceBoundaryCluster(),
				"hcp_consul_cluster":                 resourceConsulCluster(),
				"hcp_consul_cluster_root_token":      resourceConsulClusterRootToken(),
				"hcp_consul_snapshot":                resourceConsulSnapshot(),
				"hcp_hvn":                            resourceHvn(),
				"hcp_hvn_peering_connection":         resourceHvnPeeringConnection(),
				"hcp_hvn_route":                      resourceHvnRoute(),
				"hcp_packer_channel":                 resourcePackerChannel(),
				"hcp_packer_channel_assignment":      resourcePackerChannelAssignment(),
				"hcp_packer_run_task":                resourcePackerRunTask(),
				"hcp_vault_cluster":                  resourceVaultCluster(),
				"hcp_vault_cluster_admin_token":      resourceVaultClusterAdminToken(),
				"hcp_vault_plugin":                   resourceVaultPlugin(),
			},
			Schema: map[string]*schema.Schema{
				"client_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The OAuth2 Client ID for API operations.",
				},
				"client_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The OAuth2 Client Secret for API operations.",
				},
				"project_id": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.IsUUID,
					Description:  "The default project in which resources should be created.",
				},
				"credential_file": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The path to an HCP credential file to use to authenticate the provider to HCP. " +
						"You can alternatively set the HCP_CRED_FILE environment variable to point at a credential file as well. " +
						"Using a credential file allows you to authenticate the provider as a service principal via client " +
						"credentials or dynamically based on Workload Identity Federation.",
				},
				"workload_identity": {
					Type:     schema.TypeList,
					Optional: true,
					Description: "Allows authenticating the provider by exchanging the OAuth 2.0 access token or OpenID Connect " +
						"token specified in the `token_file` for a HCP service principal using Workload Identity Federation.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"token_file": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The path to a file containing a JWT token retrieved from an OpenID Connect (OIDC) or OAuth2 provider. Exactly one of `token_file` or `token` must be set.",
							},
							"token": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The JWT token retrieved from an OpenID Connect (OIDC) or OAuth2 provider. Exactly one of `token_file` or `token` must be set.",
							},
							"resource_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The resource_name of the Workload Identity Provider to exchange the token with.",
							},
						},
					},
				},
			},
			ProviderMetaSchema: map[string]*schema.Schema{
				"module_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name of the module used with the provider. Should be set in the terraform config block of the module.",
				},
			},
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

func configure(p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		var diags diag.Diagnostics
		// In order to avoid disrupting testing and development, the HCP status check only runs on prod.
		// HCP_API_HOST is used to point the provider at test environments. When unset, the provider points to prod.
		if os.Getenv("HCP_API_HOST") == "" || os.Getenv("HCP_API_HOST") == "api.cloud.hashicorp.com" {
			// This helper verifies HCP's status and either returns a warning for degraded performance
			// or errors out if there's an outage.
			diags = isHCPOperational()
		}

		clientConfig := clients.ClientConfig{
			ClientID:       d.Get("client_id").(string),
			ClientSecret:   d.Get("client_secret").(string),
			CredentialFile: d.Get("credential_file").(string),
			ProjectID:      d.Get("project_id").(string),
			SourceChannel:  p.UserAgent("terraform-provider-hcp", version.ProviderVersion),
		}

		// Read the workload_identity configuration
		if v, ok := d.GetOk("workload_identity"); ok && len(v.([]interface{})) == 1 && v.([]interface{})[0] != nil {
			wi := v.([]interface{})[0].(map[string]interface{})
			if tf, ok := wi["token_file"].(string); ok && tf != "" {
				clientConfig.WorkloadIdentityTokenFile = tf
			}
			if t, ok := wi["token"].(string); ok && t != "" {
				clientConfig.WorkloadIdentityToken = t
			}
			if rn, ok := wi["resource_name"].(string); ok && rn != "" {
				clientConfig.WorkloadIdentityResourceName = rn
			}

			if clientConfig.WorkloadIdentityTokenFile == "" && clientConfig.WorkloadIdentityToken == "" {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid workload_identity",
					Detail:        "exactly one of `token_file` or `token` must be set",
					AttributePath: cty.GetAttrPath("workload_identity"),
				})
				return nil, diags
			}
			if clientConfig.WorkloadIdentityTokenFile != "" && clientConfig.WorkloadIdentityToken != "" {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "invalid workload_identity",
					Detail:        "exactly one of `token_file` or `token` must be set",
					AttributePath: cty.GetAttrPath("workload_identity"),
				})
				return nil, diags
			}
		}

		client, err := clients.NewClient(clientConfig)
		if err != nil {
			diags = append(diags, diag.Errorf("unable to create HCP api client: %v", err)...)
			return nil, diags
		}

		// Attempt to source from the environment if unset.
		if clientConfig.ProjectID == "" {
			clientConfig.ProjectID = os.Getenv("HCP_PROJECT_ID")
		}

		if clientConfig.ProjectID != "" {
			getProjParams := project_service.NewProjectServiceGetParams()
			getProjParams.ID = clientConfig.ProjectID
			project, err := clients.RetryProjectServiceGet(client, getProjParams)
			if err != nil {
				diags = append(diags, diag.Errorf("unable to fetch project %q: %v", clientConfig.ProjectID, err)...)
				return nil, diags
			}

			client.Config.ProjectID = project.Payload.Project.ID
			client.Config.OrganizationID = project.Payload.Project.Parent.ID

		} else {
			// For the initial release of the HCP TFP, since only one project was allowed per organization at the time,
			// the provider handled used the single organization's single project by default, instead of requiring the
			// user to set it. Once multiple projects are available, this helper issues a warning: when multiple projects exist within the org,
			// a project ID should be set on the provider or on each resource. Otherwise, the oldest project will be used by default.
			// This helper will eventually be deprecated after a migration period.
			project, projDiags := getProjectFromCredentials(ctx, client)
			if projDiags != nil {
				if !projDiags.HasError() {
					diags = append(diags, projDiags...)
				} else {
					projDiags = append(projDiags, diag.Errorf("unable to get project from credentials")...)
					diags = append(diags, projDiags...)
					return nil, diags
				}
			}

			client.Config.OrganizationID = project.Parent.ID
			client.Config.ProjectID = project.ID
		}

		return client, diags
	}
}

// getProjectFromCredentials uses the configured client credentials to
// fetch the associated organization and returns that organization's
// single project.
func getProjectFromCredentials(ctx context.Context, client *clients.Client) (project *models.HashicorpCloudResourcemanagerProject, diags diag.Diagnostics) {
	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := clients.RetryOrganizationServiceList(client, listOrgParams)
	if err != nil {
		diags = append(diags, diag.Errorf("unable to fetch organization list: %v", err)...)
		return nil, diags
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The configured credentials do not have access to any organization.",
			Detail:   "Please assign at least one organization to the configured credentials to use this provider.",
		})
		return nil, diags
	}
	if orgLen > 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There is more than one organization associated with the configured credentials.",
			Detail:   "Please configure a specific project in the HCP provider config block",
		})
		return nil, diags
	}

	orgID := listOrgResp.Payload.Organizations[0].ID

	// Get the project using the organization ID.
	listProjParams := project_service.NewProjectServiceListParams()
	listProjParams.ScopeID = &orgID
	scopeType := string(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := clients.RetryProjectServiceList(client, listProjParams)
	if err != nil {
		diags = append(diags, diag.Errorf("unable to fetch project id: %v", err)...)
		return nil, diags
	}
	if len(listProjResp.Payload.Projects) > 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "There is more than one project associated with the organization of the configured credentials.",
			Detail:   `The oldest project has been selected as the default. To configure which project is used as default, set a project in the HCP provider config block. Resources may also be configured with different projects.`,
		})
		return GetOldestProject(listProjResp.Payload.Projects), diags
	}
	project = listProjResp.Payload.Projects[0]
	return project, diags
}

// GetOldestProject retrieves the oldest project from a list based on its created_at time.
func GetOldestProject(projects []*models.HashicorpCloudResourcemanagerProject) (oldestProj *models.HashicorpCloudResourcemanagerProject) {
	oldestTime := time.Now()

	for _, proj := range projects {
		projTime := time.Time(proj.CreatedAt)
		if projTime.Before(oldestTime) {
			oldestProj = proj
			oldestTime = projTime
		}
	}
	return oldestProj
}
