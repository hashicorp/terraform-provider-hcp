// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"time"

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
				"hcp_packer_image_iteration":         dataSourcePackerImageIteration(),
				"hcp_packer_image":                   dataSourcePackerImage(),
				"hcp_packer_iteration":               dataSourcePackerIteration(),
				"hcp_packer_run_task":                dataSourcePackerRunTask(),
				"hcp_vault_cluster":                  dataSourceVaultCluster(),
				"hcp_vault_secrets_app":              dataSourceVaultSecretsApp(),
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
			},
			Schema: map[string]*schema.Schema{
				"client_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_CLIENT_ID", nil),
					Description: "The OAuth2 Client ID for API operations.",
				},
				"client_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_CLIENT_SECRET", nil),
					Description: "The OAuth2 Client Secret for API operations.",
				},
				"project_id": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("HCP_PROJECT_ID", nil),
					ValidateFunc: validation.IsUUID,
					Description:  "The default project in which resources should be created.",
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

		// Sets up HCP SDK client.
		userAgent := p.UserAgent("terraform-provider-hcp", version.ProviderVersion)
		clientID := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)

		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			SourceChannel: userAgent,
		})
		if err != nil {
			diags = append(diags, diag.Errorf("unable to create HCP api client: %v", err)...)
			return nil, diags
		}

		projectID := d.Get("project_id").(string)

		if projectID != "" {
			getProjParams := project_service.NewProjectServiceGetParams()
			getProjParams.ID = projectID
			project, err := clients.RetryProjectServiceGet(client, getProjParams)
			if err != nil {
				diags = append(diags, diag.Errorf("unable to fetch project %q: %v", projectID, err)...)
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
func getProjectFromCredentials(ctx context.Context, client *clients.Client) (project *models.ResourcemanagerProject, diags diag.Diagnostics) {
	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := clients.RetryOrganizationServiceList(client, listOrgParams)
	if err != nil {
		diags = append(diags, diag.Errorf("unable to fetch organization list: %v", err)...)
		return nil, diags
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen != 1 {
		diags = append(diags, diag.Errorf("unexpected number of organizations: expected 1, actual: %v", orgLen)...)
		return nil, diags
	}
	orgID := listOrgResp.Payload.Organizations[0].ID

	// Get the project using the organization ID.
	listProjParams := project_service.NewProjectServiceListParams()
	listProjParams.ScopeID = &orgID
	scopeType := string(models.ResourceIDResourceTypeORGANIZATION)
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
		return getOldestProject(listProjResp.Payload.Projects), diags
	}
	project = listProjResp.Payload.Projects[0]
	return project, diags
}

// getOldestProject retrieves the oldest project from a list based on its created_at time.
func getOldestProject(projects []*models.ResourcemanagerProject) (oldestProj *models.ResourcemanagerProject) {
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
