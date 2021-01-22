package provider

import (
	"context"
	"log"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
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
				"organization_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_ORGANIZATION_ID", ""),
					Description: "The ID of the organization for API operations.",
				},
				"project_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("HCP_PROJECT_ID", ""),
					Description: "The ID of the project for API operations.",
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

		// TODO: can i fetch org and project IDs here?
		// use an initial client to fetch org ID, then proj ID,
		if d.Get("organization_id").(string) == "" || d.Get("project_id").(string) == "" {

			// TODO: drop debug prints
			log.Printf("[HEY!!!] Initializing Client with id %s and secret %s", d.Get("client_id"), d.Get("client_secret"))
			cl, err := clients.NewClient(clients.ClientConfig{
				ClientID:     d.Get("client_id").(string),
				ClientSecret: d.Get("client_secret").(string),
			})
			if err != nil {
				return nil, diag.Errorf("unable to create HCP api client: %v", err)
			}

			listOrgParams := organization_service.NewOrganizationServiceListParams()
			listOrgResp, err := cl.Organization.OrganizationServiceList(listOrgParams, nil)
			log.Printf("[HEY!!!] Org List Resp %+v", listOrgResp)
			if err != nil {
				return nil, diag.Errorf("unable to fetch organization list: %+v", err)
			}

			orgID := listOrgResp.Payload.Organizations[0].ID
			log.Printf("[HEY!!!] Org ID %s", orgID)

			if err := d.Set("organization_id", orgID); err != nil {
				return nil, diag.Errorf("unable to set organization_id: %+v", err)
			}

			listProjParams := project_service.NewProjectServiceListParams()
			listProjParams.ScopeID = &orgID
			scopeType := "ORGANIZATION"
			listProjParams.ScopeType = &scopeType

			listProjResp, err := cl.Project.ProjectServiceList(listProjParams, nil)
			log.Printf("[HEY!!!] Proj List Resp %+v", listProjResp)
			if err != nil {
				return nil, diag.Errorf("unable to fetch project list: %+v", err)
			}

			if len(listProjResp.Payload.Projects) > 1 {
				return nil, diag.Errorf("More than one project detected. This version of the provider does not support multiple projects.")
			}

			projID := listProjResp.Payload.Projects[0].ID
			log.Printf("[HEY!!!] Proj ID %s", projID)

			if err := d.Set("project_id", projID); err != nil {
				return nil, diag.Errorf("unable to set project_id: %+v", err)
			}
		}

		// Construct a new HCP api client with clients and configuration.
		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:       d.Get("client_id").(string),
			ClientSecret:   d.Get("client_secret").(string),
			OrganizationID: d.Get("organization_id").(string),
			ProjectID:      d.Get("project_id").(string),
			SourceChannel:  userAgent,
		})
		if err != nil {
			return nil, diag.Errorf("unable to create fully-configured HCP api client: %v", err)
		}

		return client, nil
	}
}
