package provider

import (
	"context"
	"log"

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

		project, err := getProject(ctx, clientID, clientSecret)
		if err != nil {
			// TODO figure out how to bubble up this error
			return nil, diag.Errorf("unable to determine project and organization ID from configured credentials")
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

// return pointer to HashicorpCloudResourcemanagerProject and err (2 tuple return)
// HashicorpCloudResourcemanagerResourceID is the org id because only orgs are allowed to be project parents currently, although it may change in the future
func getProject(ctx context.Context, clientID string, clientSecret string) (*models.HashicorpCloudResourcemanagerProject, error) {
	// func getProject(ctx context.Context, clientID string, clientSecret string) {
	log.Printf("**********")
	log.Printf("in getProject")
	log.Printf("**********")

	// if d.Get("organization_id").(string) == "" {

	// TODO: drop debug prints
	log.Printf("**********")
	log.Printf("[HEY!!!] Initializing Client with id %s and secret %s", clientID, clientSecret)
	log.Printf("**********")

	cl, _ := clients.NewClient(clients.ClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	// if err != nil {
	// 	return diag.Errorf("unable to create HCP api client: %v", err)
	// }

	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, _ := cl.Organization.OrganizationServiceList(listOrgParams, nil)
	log.Printf("**********")
	log.Printf("[HEY!!!] Org List Resp %+v", listOrgResp)
	log.Printf("**********")

	// if err != nil {
	// 	return diag.Errorf("unable to fetch organization list: %+v", err)
	// }

	// Service principles, from which the client credentials are obtained, are scoped to a single org,
	// so this list should never have more than one org.
	orgID := listOrgResp.Payload.Organizations[0].ID
	log.Printf("**********")
	log.Printf("[HEY!!!] Org ID %s", orgID)
	log.Printf("**********")

	// if err := d.Set("organization_id", orgID); err != nil {
	// 	return diag.Errorf("unable to set organization_id: %+v", err)
	// }
	// }
	listProjParams := project_service.NewProjectServiceListParams()
	listProjParams.ScopeID = &orgID
	scopeType := "ORGANIZATION" // TODO use the enum
	listProjParams.ScopeType = &scopeType

	listProjResp, _ := cl.Project.ProjectServiceList(listProjParams, nil)
	log.Printf("**********")
	log.Printf("[HEY!!!] Proj List Resp %+v", listProjResp)
	log.Printf("**********")

	project := listProjResp.Payload.Projects[0]
	log.Printf("**********")
	log.Printf("[HEY!!!] Project %v", project)
	log.Printf("**********")

	projID := listProjResp.Payload.Projects[0].ID
	log.Printf("**********")
	log.Printf("[HEY!!!] Proj ID %s", projID)
	log.Printf("**********")

	return project, nil

	// if err != nil {
	// 	return nil, diag.Errorf("unable to fetch project list: %+v", err)
	// }

	// return diagnostics
}
