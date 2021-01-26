package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"hcp_consul_cluster": dataSourceConsulCluster(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hcp_aws_network_peering":       resourceAwsNetworkPeering(),
				"hcp_consul_cluster":            resourceConsulCluster(),
				"hcp_consul_cluster_root_token": resourceConsulClusterRootToken(),
				"hcp_hvn":                       resourceHvn(),
			},
			Schema: map[string]*schema.Schema{
				"client_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_CLIENT_ID", nil),
					Description: "The OAuth2 Client ID for API operations.",
				},
				"client_secret": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_CLIENT_SECRET", nil),
					Description: "The OAuth2 Client Secret for API operations.",
				},
			},
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

func configure(p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-hcp", version.ProviderVersion)
		clientID := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)

		// For the initial release, since only one project is allowed per organization, the
		// provider handles fetching the organization's single project, instead of requiring the
		// user to set it. When multiple projects are supported, this helper will be deprecated
		// with a warning: when multiple projects exist within the org, a project ID must be set
		// on the provider or on each resource.
		project, err := getProjectFromCredentials(ctx, clientID, clientSecret)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Construct a new HCP api client with clients and configuration.
		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			OrganizationID: project.Parent.ID,
			ProjectID:      project.ID,
			SourceChannel:  userAgent,
		})
		if err != nil {
			return nil, diag.Errorf("unable to create HCP api client: %v", err)
		}

		return client, nil
	}
}

// getProjectFromCredentials uses the configured client credentials to
//  fetch the associated organization and returns that organization's
// single project.
func getProjectFromCredentials(ctx context.Context, clientID string, clientSecret string) (*models.HashicorpCloudResourcemanagerProject, error) {
	// Create a client to use for querying organization.
	cl, err := clients.NewClient(clients.ClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create HCP api client: %v", err)
	}

	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := cl.Organization.OrganizationServiceList(listOrgParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch organization list: %+v", err)
	}
	orgLen := len(listOrgResp.Payload.Organizations)
	if orgLen != 1 {
		return nil, fmt.Errorf("unexpected number of organizations: expected 1, actual: %v", orgLen)
	}
	orgID := listOrgResp.Payload.Organizations[0].ID

	// Get the project using the organization ID.
	listProjParams := project_service.NewProjectServiceListParams()
	listProjParams.ScopeID = &orgID
	scopeType := string(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listProjParams.ScopeType = &scopeType
	listProjResp, err := cl.Project.ProjectServiceList(listProjParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch project id: %+v", err)
	}
	if len(listProjResp.Payload.Projects) > 1 {
		return nil, fmt.Errorf("this version of the provider does not support multiple projects. Upgrade to a later provider version and set a project ID on the provider/resources.")
	}

	project := listProjResp.Payload.Projects[0]
	return project, nil
}
