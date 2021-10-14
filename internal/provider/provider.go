package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/preview/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/version"
)

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"hcp_aws_network_peering":            dataSourceAwsNetworkPeering(),
				"hcp_aws_transit_gateway_attachment": dataSourceAwsTransitGatewayAttachment(),
				"hcp_consul_agent_helm_config":       dataSourceConsulAgentHelmConfig(),
				"hcp_consul_agent_kubernetes_secret": dataSourceConsulAgentKubernetesSecret(),
				"hcp_consul_cluster":                 dataSourceConsulCluster(),
				"hcp_consul_versions":                dataSourceConsulVersions(),
				"hcp_hvn":                            dataSourceHvn(),
				"hcp_hvn_peering_connection":         dataSourceHvnPeeringConnection(),
				"hcp_hvn_route":                      dataSourceHVNRoute(),
				"hcp_vault_cluster":                  dataSourceVaultCluster(),
				"hcp_packer_image_iteration":         dataSourcePackerImageIteration(),
				"hcp_packer_image":                   dataSourcePackerImage(),
				"hcp_packer_iteration":               dataSourcePackerIteration(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hcp_aws_network_peering":            resourceAwsNetworkPeering(),
				"hcp_aws_transit_gateway_attachment": resourceAwsTransitGatewayAttachment(),
				"hcp_consul_cluster":                 resourceConsulCluster(),
				"hcp_consul_cluster_root_token":      resourceConsulClusterRootToken(),
				"hcp_consul_snapshot":                resourceConsulSnapshot(),
				"hcp_hvn":                            resourceHvn(),
				"hcp_hvn_peering_connection":         resourceHvnPeeringConnection(),
				"hcp_hvn_route":                      resourceHvnRoute(),
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

		// For the initial release, since only one project is allowed per organization, the
		// provider handles fetching the organization's single project, instead of requiring the
		// user to set it. When multiple projects are supported, this helper will be deprecated
		// with a warning: when multiple projects exist within the org, a project ID must be set
		// on the provider or on each resource.
		project, err := getProjectFromCredentials(ctx, client)
		if err != nil {
			diags = append(diags, diag.Errorf("unable to get project from credentials: %v", err)...)
			return nil, diags
		}
		client.Config.OrganizationID = project.Parent.ID
		client.Config.ProjectID = project.ID

		return client, diags
	}
}

// getProjectFromCredentials uses the configured client credentials to
// fetch the associated organization and returns that organization's
// single project.
func getProjectFromCredentials(ctx context.Context, client *clients.Client) (*models.HashicorpCloudResourcemanagerProject, error) {
	// Get the organization ID.
	listOrgParams := organization_service.NewOrganizationServiceListParams()
	listOrgResp, err := client.Organization.OrganizationServiceList(listOrgParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch organization list: %v", err)
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
	listProjResp, err := client.Project.ProjectServiceList(listProjParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch project id: %v", err)
	}
	if len(listProjResp.Payload.Projects) > 1 {
		return nil, fmt.Errorf("this version of the provider does not support multiple projects, upgrade to a later provider version and set a project ID on the provider/resources")
	}

	project := listProjResp.Payload.Projects[0]
	return project, nil
}

// Status endpoint for prod.
const statuspageUrl = "https://pdrzb3d64wsj.statuspage.io/api/v2/components.json"
const statuspageHcpComponentId = "ym75hzpmfq4q"

type status string

// Possible statuses returned by statuspage.io.
const (
	operational         status = "operational"
	degradedPerformance        = "degraded_performance"
	partialOutage              = "partial_outage"
	majorOutage                = "major_outage"
	underMaintenance           = "under_maintenance"
)

type statuspage struct {
	Components []component `json:"components"`
}

type component struct {
	ID     string `json:"id"`
	Status status `json:"status"`
}

func isHCPOperational() diag.Diagnostics {
	req, err := http.NewRequest("GET", statuspageUrl, nil)
	if err != nil {
		log.Printf("Unable to create request to verify HCP status: %s", err)
	}

	var cl = http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		log.Printf("Unable complete request to verify HCP status: %s", err)
	}
	defer resp.Body.Close()

	jsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Unable read response to verify HCP status: %s", err)
	}

	sp := statuspage{}
	err = json.Unmarshal(jsBytes, &sp)
	if err != nil {
		log.Printf("Unable unmarshal response to verify HCP status: %s", err)
	}

	var st status
	for _, c := range sp.Components {
		if c.ID == statuspageHcpComponentId {
			st = c.Status
		}
	}

	var diags diag.Diagnostics

	switch st {
	case operational:
		log.Printf("HCP is fully operational.")
	case partialOutage, majorOutage:
		return diag.Errorf("HCP is experiencing an outage. Please check https://status.hashicorp.com for more details.")
	case degradedPerformance:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "HCP is experiencing degraded performance.",
			Detail:   "Please check https://status.hashicorp.com for more details.",
		})
	case underMaintenance:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "HCP is undergoing maintenance that may affect performance.",
			Detail:   "Please check https://status.hashicorp.com for more details.",
		})
	}

	return diags

}
