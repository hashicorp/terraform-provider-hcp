package provider

import (
	"context"
	"log"

	"github.com/hashicorp/cloud-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
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

		orgDiag := ensureOrg(ctx, d)
		if orgDiag.HasError() {
			return nil, diag.Errorf("unable to set organization ID from configured credentials")
		}

		// Construct a new HCP api client with clients and configuration.
		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:       d.Get("client_id").(string),
			ClientSecret:   d.Get("client_secret").(string),
			OrganizationID: d.Get("organization_id").(string),
			ProjectID:      d.Get("project_id").(string), // can be blank if set on the resource
			SourceChannel:  userAgent,
		})
		if err != nil {
			return nil, diag.Errorf("unable to create HCP api client: %v", err)
		}

		return client, nil
	}
}

func ensureOrg(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if d.Get("organization_id").(string) == "" {

		// TODO: drop debug prints
		log.Printf("[HEY!!!] Initializing Client with id %s and secret %s", d.Get("client_id"), d.Get("client_secret"))
		cl, err := clients.NewClient(clients.ClientConfig{
			ClientID:     d.Get("client_id").(string),
			ClientSecret: d.Get("client_secret").(string),
		})
		if err != nil {
			return diag.Errorf("unable to create HCP api client: %v", err)
		}

		listOrgParams := organization_service.NewOrganizationServiceListParams()
		listOrgResp, err := cl.Organization.OrganizationServiceList(listOrgParams, nil)
		log.Printf("[HEY!!!] Org List Resp %+v", listOrgResp)
		if err != nil {
			return diag.Errorf("unable to fetch organization list: %+v", err)
		}

		// Service principles, from which the client credentials are obtained, are scoped to a single org,
		// so this list should never have more than one org.
		orgID := listOrgResp.Payload.Organizations[0].ID
		log.Printf("[HEY!!!] Org ID %s", orgID)

		if err := d.Set("organization_id", orgID); err != nil {
			return diag.Errorf("unable to set organization_id: %+v", err)
		}
	}

	return diagnostics
}
