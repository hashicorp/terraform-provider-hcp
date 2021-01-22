package provider

import (
	"context"

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
				"hcp_consul_snapshot":           resourceConsulSnapshot(),
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

		// Construct a new HCP api client with clients and configuration.
		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:       d.Get("client_id").(string),
			ClientSecret:   d.Get("client_secret").(string),
			OrganizationID: d.Get("organization_id").(string),
			ProjectID:      d.Get("project_id").(string),
			SourceChannel:  userAgent,
		})
		if err != nil {
			return nil, diag.Errorf("unable to create HCP api client: %v", err)
		}

		return client, nil
	}
}
