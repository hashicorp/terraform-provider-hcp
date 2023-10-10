// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	diagnostic "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"

	"github.com/hashicorp/terraform-provider-hcp/internal/provider/resourcemanager"
)

// This is an implementation using the Provider framework
// Docs can be found here: https://developer.hashicorp.com/terraform/plugin/framework
// NOTE: All other resources and data sources for other products can be found in the
// providersdkv2 folder at the same level
type ProviderFramework struct {
	version string
}

type ProviderFrameworkConfiguration struct {
	Client *clients.Client
}

type ProviderFrameworkModel struct {
	ClientSecret types.String `tfsdk:"client_secret"`
	ClientID     types.String `tfsdk:"client_id"`
	ProjectID    types.String `tfsdk:"project_id"`
}

func (p *ProviderFramework) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hcp"
	resp.Version = "dev"
}

func (p *ProviderFramework) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth2 Client ID for API operations.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth2 Client Secret for API operations.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The default project in which resources should be created.",
			},
		},
	}
}

func (p *ProviderFramework) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVaultSecretsAppResource,
		NewVaultSecretsSecretResource,
		resourcemanager.NewProjectResource,
	}
}

func (p *ProviderFramework) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVaultSecretsAppDataSource,
		NewVaultSecretsSecretDataSource,
		resourcemanager.NewProjectDataSource,
		resourcemanager.NewOrganizationDataSource,
	}
}

func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProviderFramework{
			version: version,
		}
	}
}

func (p *ProviderFramework) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// In order to avoid disrupting testing and development, the HCP status check only runs on prod.
	// HCP_API_HOST is used to point the provider at test environments. When unset, the provider points to prod.
	if os.Getenv("HCP_API_HOST") == "" || os.Getenv("HCP_API_HOST") == "api.cloud.hashicorp.com" {
		// This helper verifies HCP's status and either returns a warning for degraded performance
		// or errors out if there's an outage.
		resp.Diagnostics.Append(isHCPOperationalFramework()...)
	}

	// Sets up HCP SDK client.
	var data ProviderFrameworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	clientID := ""
	if data.ClientID.ValueString() != "" {
		clientID = data.ClientID.ValueString()
	} else {
		clientID = os.Getenv("HCP_CLIENT_ID")
	}

	clientSecret := ""
	if data.ClientSecret.ValueString() != "" {
		clientSecret = data.ClientSecret.ValueString()
	} else {
		clientSecret = os.Getenv("HCP_CLIENT_SECRET")
	}

	client, err := clients.NewClient(clients.ClientConfig{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		SourceChannel: "terraform-provider-hcp",
	})

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("unable to create HCP api client: %v", err), "")
		return
	}

	projectID := ""
	if data.ProjectID.ValueString() != "" {
		projectID = data.ProjectID.ValueString()
	} else {
		projectID = os.Getenv("HCP_PROJECT_ID")
	}

	if projectID != "" {
		getProjParams := project_service.NewProjectServiceGetParams()
		getProjParams.ID = projectID
		project, err := clients.RetryProjectServiceGet(client, getProjParams)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("unable to fetch project %q: %v", projectID, err), "")
			return
		}

		client.Config.ProjectID = project.Payload.Project.ID
		client.Config.OrganizationID = project.Payload.Project.Parent.ID

	} else {
		// For the initial release of the HCP TFP, since only one project was allowed per organization at the time,
		// the provider handled used the single organization's single project by default, instead of requiring the
		// user to set it. Once multiple projects are available, this helper issues a warning: when multiple projects exist within the org,
		// a project ID should be set on the provider or on each resource. Otherwise, the oldest project will be used by default.
		// This helper will eventually be deprecated after a migration period.
		project, projDiags := getProjectFromCredentialsFramework(ctx, client)
		if projDiags != nil {
			if !projDiags.HasError() {
				resp.Diagnostics.Append(projDiags...)
			} else {
				resp.Diagnostics.AddError("unable to get project from credentials", "")
				return
			}
		}

		client.Config.OrganizationID = project.Parent.ID
		client.Config.ProjectID = project.ID
	}

	var config ProviderFrameworkConfiguration
	config.Client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

// getProjectFromCredentials uses the configured client credentials to
// fetch the associated organization and returns that organization's
// single project.
// This differs from the provider.go implementation due to the diagnostics used
// by the plugin framework.
func getProjectFromCredentialsFramework(ctx context.Context, client *clients.Client) (project *models.HashicorpCloudResourcemanagerProject, diags diagnostic.Diagnostics) {
	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := clients.RetryOrganizationServiceList(client, listOrgParams)
	if err != nil {
		diags.AddError(fmt.Sprintf("unable to fetch organization list: %v", err), "")

		return nil, diags
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen != 1 {
		diags.AddError(fmt.Sprintf("unexpected number of organizations: expected 1, actual: %v", orgLen), "")
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
		diags.AddError(fmt.Sprintf("unable to fetch project id: %v", err), "")
		return nil, diags
	}
	if len(listProjResp.Payload.Projects) > 1 {
		diags.AddWarning("There is more than one project associated with the organization of the configured credentials.", `The oldest project has been selected as the default. To configure which project is used as default, set a project in the HCP provider config block. Resources may also be configured with different projects.`)

		return getOldestProject(listProjResp.Payload.Projects), diags
	}
	project = listProjResp.Payload.Projects[0]
	return project, diags
}

// getOldestProject retrieves the oldest project from a list based on its created_at time.
func getOldestProject(projects []*models.HashicorpCloudResourcemanagerProject) (oldestProj *models.HashicorpCloudResourcemanagerProject) {
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
