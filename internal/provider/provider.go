package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

type statuspage struct {
	Components []component `json:"components"`
}

type status string

type component struct {
	ID     string `json:"id"`
	Status status `json:"status"`
}

const statuspageUrl = "https://pdrzb3d64wsj.statuspage.io/api/v2/components.json"
const statuspageHcpComponentId = "ym75hzpmfq4q"

// Possible statuses returned by statuspage.io.
const (
	operational         status = "operational"
	degradedPerformance        = "degraded_performance"
	partialOutage              = "partial_outage"
	majorOutage                = "major_outage"
	underMaintenance           = "under_maintenance"
)

func configure(p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		js := `{"page":{"id":"pdrzb3d64wsj","name":"HashiCorp Services","url":"https://status.hashicorp.com","time_zone":"Etc/UTC","updated_at":"2021-08-06T23:51:55.230Z"},"components":[{"id":"7cgbxrw10wgr","name":"Terraform Cloud UI and API","status": "degraded_performance","created_at":"2015-06-15T21:22:42.700Z","updated_at":"2021-06-25T22:39:19.797Z","position":1,"description":"Frontend, APIs, UI, help and public website","showcase":true,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"gf9n1hc7qvs0","name":"Running Consul clusters","status": "degraded_performance","created_at":"2020-07-22T22:24:28.497Z","updated_at":"2020-07-22T22:24:28.497Z","position":1,"description":"The availability and health of Consul clusters provisioned via HCS","showcase":false,"start_date":null,"group_id":"9wcfddfz74dn","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"cpn6w07r5f2y","name":"Waypoint URL Service","status": "degraded_performance","created_at":"2020-10-16T17:34:46.558Z","updated_at":"2021-01-29T02:33:36.520Z","position":1,"description":null,"showcase":false,"start_date":"2020-10-16","group_id":"j0bgx9v6fcp9","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"ym75hzpmfq4q","name":"HashiCorp Cloud Platform (HCP)","status": "under_maintenance","created_at":"2021-01-14T17:03:37.911Z","updated_at":"2021-02-23T17:53:38.043Z","position":1,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["mgv1p2j9x444","b4pr9x7288zy","c32yfw32879v"]},{"id":"vm6n1x4n30pw","name":"Releases","status": "degraded_performance","created_at":"2021-05-13T19:58:53.911Z","updated_at":"2021-05-13T19:58:53.911Z","position":1,"description":"Official downloads site for HashiCorp product binaries (releases.hashicorp.com)","showcase":false,"start_date":"2021-05-13","group_id":"v9vx6ynhfdwd","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"mgv1p2j9x444","name":"HCP Portal","status": "degraded_performance","created_at":"2021-01-14T17:03:37.921Z","updated_at":"2021-07-28T22:17:46.579Z","position":1,"description":"Web portal, enabling creation, deletion and operation of managed clusters in HCP.","showcase":true,"start_date":"2021-01-14","group_id":"ym75hzpmfq4q","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"l8mnm91h3s7v","name":"Terraform Registry","status": "degraded_performance","created_at":"2017-09-29T18:00:10.681Z","updated_at":"2021-07-17T01:22:02.945Z","position":1,"description":"Publishing and discovery of community Terraform modules\r\n\r\nRelies on the Fastly - https://status.fastly.com/","showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"b4pr9x7288zy","name":"Consul","status": "degraded_performance","created_at":"2021-01-14T17:04:09.051Z","updated_at":"2021-07-19T21:22:13.294Z","position":2,"description":"Health of deployed Consul clusters, managed by HCP.","showcase":false,"start_date":"2021-01-14","group_id":"ym75hzpmfq4q","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"c5017zlm6qhz","name":"Terraform Runs","status": "degraded_performance","created_at":"2015-06-17T19:21:42.598Z","updated_at":"2021-07-22T15:01:00.960Z","position":2,"description":"Workers that run Terraform","showcase":true,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"mgkn9n8p7y6g","name":"Heroku","status": "degraded_performance","created_at":"2020-03-16T19:15:52.001Z","updated_at":"2020-03-16T19:57:05.957Z","position":2,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"9m881ktq4lfx","name":"Creation via Azure Marketplace","status": "degraded_performance","created_at":"2020-07-22T22:25:28.097Z","updated_at":"2021-07-28T22:17:46.598Z","position":2,"description":"Provisioning new Consul clusters via the Azure Marketplace","showcase":false,"start_date":null,"group_id":"9wcfddfz74dn","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"cnzk9zc90jqy","name":"Terraform Cloud","status":"under_maintenance","created_at":"2019-03-26T17:29:41.837Z","updated_at":"2021-01-14T17:10:48.846Z","position":2,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["7cgbxrw10wgr","c5017zlm6qhz","fd91hrtfhxp2","s8ypd7vzf954","gl5yfxj7t9r8","n5znzfs17thw","877cyd1dq17r","96dyvdwnj91k","dq4zxvvplkk9","ng26sgmqxvfj","gq1111m9fzlb","htt5nclyn7c0","xtp606778253","tzr6rb1xj9b6","hyg0p6bsnm3w","ngczmzdl26k7","wwjttb02dfks","c232669hd65f","k2qq7zh1hmr3","5n7htt0nqfjn","m9gbmx8cnhtz","rwq0p2v3vj2n","kccwgg563fsk","jsn2mgtr3frt","c5dh96zvq36s","61srm3006btq","4wpn4b8jhcr3","sws7g0djf9zz","flmmxyq8hyjc","8f1p0crvqs3t","wxkqsbv24jp9","m75hplc2vpvl"]},{"id":"5dwc7bq1bxtl","name":"Deletion of Consul clusters","status": "degraded_performance","created_at":"2020-07-22T22:26:07.410Z","updated_at":"2021-07-28T22:17:46.618Z","position":3,"description":"Deleting Azure Managed Apps via the Azure portal","showcase":false,"start_date":null,"group_id":"9wcfddfz74dn","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"fd91hrtfhxp2","name":"Private Module Registry","status": "degraded_performance","created_at":"2019-03-26T18:07:20.520Z","updated_at":"2021-05-27T08:21:28.349Z","position":3,"description":null,"showcase":true,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"0jc2px93ctvg","name":"Terraform Registry","status": "degraded_performance","created_at":"2020-03-16T19:16:33.472Z","updated_at":"2021-06-25T17:43:58.926Z","position":3,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["l8mnm91h3s7v","mgkn9n8p7y6g","jbrwb9tvgfqs","6flgq3bdl399","78rbrv1f6sdr","pmmv4f6dgxvt","qyy7zy00w3sx"]},{"id":"s8ypd7vzf954","name":"Cost Estimation","status": "degraded_performance","created_at":"2020-01-22T00:53:55.436Z","updated_at":"2021-05-27T08:21:28.377Z","position":4,"description":null,"showcase":true,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"c32yfw32879v","name":"Vault","status": "degraded_performance","created_at":"2021-04-07T15:59:51.325Z","updated_at":"2021-07-19T21:22:13.315Z","position":4,"description":"Health of deployed Vault clusters, managed by HCP.","showcase":false,"start_date":"2021-04-07","group_id":"ym75hzpmfq4q","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"jbrwb9tvgfqs","name":"Fastly API","status": "degraded_performance","created_at":"2020-03-16T19:29:21.932Z","updated_at":"2021-07-15T20:57:03.252Z","position":4,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"1mdm36t0fkx1","name":"Vagrant Cloud","status": "degraded_performance","created_at":"2017-05-31T17:37:49.591Z","updated_at":"2021-06-25T17:44:24.591Z","position":4,"description":"Vagrant box publishing, hosting, and management","showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"9wcfddfz74dn","name":"HashiCorp Consul Service (HCS) on Azure","status": "degraded_performance","created_at":"2020-07-22T22:24:28.477Z","updated_at":"2021-06-25T17:44:24.602Z","position":5,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["gf9n1hc7qvs0","9m881ktq4lfx","5dwc7bq1bxtl"]},{"id":"6flgq3bdl399","name":"Fastly Purging","status": "degraded_performance","created_at":"2020-03-16T19:29:27.833Z","updated_at":"2021-01-05T21:19:38.093Z","position":5,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"gl5yfxj7t9r8","name":"Logging","status": "degraded_performance","created_at":"2015-06-17T19:21:53.605Z","updated_at":"2020-11-18T23:50:36.399Z","position":5,"description":"Logging for Terraform, Packer","showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"n5znzfs17thw","name":"File Storage","status": "degraded_performance","created_at":"2015-06-17T19:22:26.485Z","updated_at":"2020-11-18T23:50:36.408Z","position":6,"description":"File uploads and downloads","showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"j0bgx9v6fcp9","name":"Waypoint","status": "degraded_performance","created_at":"2020-10-16T17:34:46.549Z","updated_at":"2021-06-25T17:44:24.610Z","position":6,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["cpn6w07r5f2y"]},{"id":"78rbrv1f6sdr","name":"Fastly Configuration Deployment","status": "degraded_performance","created_at":"2020-03-16T19:29:31.042Z","updated_at":"2020-10-24T20:20:17.523Z","position":6,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"v9vx6ynhfdwd","name":"Release Archives","status": "degraded_performance","created_at":"2021-05-13T19:58:53.896Z","updated_at":"2021-06-25T17:44:24.618Z","position":7,"description":null,"showcase":false,"start_date":null,"group_id":null,"page_id":"pdrzb3d64wsj","group":true,"only_show_if_degraded":false,"components":["vm6n1x4n30pw"]},{"id":"877cyd1dq17r","name":"Atlassian Bitbucket Source downloads","status": "degraded_performance","created_at":"2020-11-18T23:54:09.442Z","updated_at":"2021-08-02T19:03:41.856Z","position":7,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"pmmv4f6dgxvt","name":"Fastly San Jose (SJC)","status": "degraded_performance","created_at":"2021-06-29T20:26:44.504Z","updated_at":"2021-07-15T15:45:47.384Z","position":7,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"96dyvdwnj91k","name":"Atlassian Bitbucket API","status":"under_maintenance","created_at":"2020-11-18T23:54:00.372Z","updated_at":"2021-08-06T20:01:08.250Z","position":8,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"qyy7zy00w3sx","name":"Fastly Singapore (SIN)","status": "degraded_performance","created_at":"2021-06-29T20:27:23.805Z","updated_at":"2021-06-29T20:28:52.470Z","position":8,"description":null,"showcase":false,"start_date":null,"group_id":"0jc2px93ctvg","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"dq4zxvvplkk9","name":"Atlassian Bitbucket Webhooks","status": "degraded_performance","created_at":"2018-08-01T16:17:14.517Z","updated_at":"2021-08-05T18:52:48.863Z","position":9,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"ng26sgmqxvfj","name":"Atlassian Bitbucket Git via HTTPS","status":"under_maintenance","created_at":"2018-08-01T16:17:23.355Z","updated_at":"2021-08-06T20:00:51.868Z","position":10,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"gq1111m9fzlb","name":"Atlassian Bitbucket SSH","status":"under_maintenance","created_at":"2018-08-01T16:17:37.084Z","updated_at":"2021-08-06T20:00:54.207Z","position":11,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"htt5nclyn7c0","name":"Twilio Outgoing SMS","status": "degraded_performance","created_at":"2015-09-02T14:03:21.684Z","updated_at":"2021-07-28T10:01:17.886Z","position":12,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"xtp606778253","name":"Twilio REST API","status": "degraded_performance","created_at":"2015-09-02T14:03:22.637Z","updated_at":"2021-08-03T08:47:57.544Z","position":13,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"tzr6rb1xj9b6","name":"Quay.io API","status": "degraded_performance","created_at":"2019-03-11T20:40:31.955Z","updated_at":"2021-03-23T04:45:19.606Z","position":14,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"hyg0p6bsnm3w","name":"Quay.io Build System","status": "degraded_performance","created_at":"2019-03-11T20:40:33.467Z","updated_at":"2021-07-26T13:24:32.489Z","position":15,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"ngczmzdl26k7","name":"Quay.io Registry","status": "degraded_performance","created_at":"2019-03-11T20:40:34.886Z","updated_at":"2021-03-23T04:45:20.233Z","position":16,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"wwjttb02dfks","name":"GitHub","status": "degraded_performance","created_at":"2015-06-18T21:22:50.046Z","updated_at":"2021-06-02T09:29:44.441Z","position":17,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"c232669hd65f","name":"GitHub Webhooks","status": "degraded_performance","created_at":"2020-11-18T23:44:38.999Z","updated_at":"2021-07-29T21:08:34.062Z","position":18,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"k2qq7zh1hmr3","name":"GitHub Notifications","status": "degraded_performance","created_at":"2019-05-22T17:38:49.759Z","updated_at":"2020-11-18T23:55:43.574Z","position":19,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"5n7htt0nqfjn","name":"GitHub API Requests","status": "degraded_performance","created_at":"2019-05-22T17:40:44.715Z","updated_at":"2020-11-18T23:55:43.579Z","position":20,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"m9gbmx8cnhtz","name":"GitHub Issues, PRs, Dashboard, Projects","status": "degraded_performance","created_at":"2019-05-22T17:44:02.625Z","updated_at":"2021-06-14T10:53:54.155Z","position":21,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"rwq0p2v3vj2n","name":"Stripe API","status": "degraded_performance","created_at":"2019-07-10T17:01:21.645Z","updated_at":"2021-07-27T16:22:14.896Z","position":22,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"kccwgg563fsk","name":"Stripe Dashboard","status": "degraded_performance","created_at":"2019-07-10T17:01:22.790Z","updated_at":"2021-07-19T17:58:39.030Z","position":23,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"jsn2mgtr3frt","name":"Stripe JS","status": "degraded_performance","created_at":"2019-07-10T17:01:23.614Z","updated_at":"2021-06-08T11:03:55.551Z","position":24,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"c5dh96zvq36s","name":"Stripe Emails","status": "degraded_performance","created_at":"2019-07-10T17:01:24.576Z","updated_at":"2020-11-18T23:55:43.608Z","position":25,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"61srm3006btq","name":"Stripe Webhooks","status": "degraded_performance","created_at":"2019-07-10T17:01:25.513Z","updated_at":"2020-11-18T23:55:43.614Z","position":26,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"4wpn4b8jhcr3","name":"AWS ec2-us-east-1","status": "degraded_performance","created_at":"2015-08-10T19:19:09.667Z","updated_at":"2021-04-13T22:05:20.024Z","position":27,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"sws7g0djf9zz","name":"AWS IAM","status": "degraded_performance","created_at":"2020-09-17T00:31:31.870Z","updated_at":"2021-06-18T00:39:32.072Z","position":28,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"flmmxyq8hyjc","name":"AWS apigateway-us-east-1","status": "degraded_performance","created_at":"2020-11-25T16:38:18.809Z","updated_at":"2020-11-25T16:39:01.000Z","position":29,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"8f1p0crvqs3t","name":"Fastly Tokyo (TYO)","status": "degraded_performance","created_at":"2021-06-25T17:41:31.485Z","updated_at":"2021-06-25T17:45:05.064Z","position":31,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"wxkqsbv24jp9","name":"Fastly London (LON)","status": "degraded_performance","created_at":"2021-06-25T17:42:10.153Z","updated_at":"2021-06-25T17:45:39.053Z","position":32,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false},{"id":"m75hplc2vpvl","name":"Fastly New York (LGA)","status": "degraded_performance","created_at":"2021-06-25T17:43:16.237Z","updated_at":"2021-06-25T17:45:54.676Z","position":33,"description":null,"showcase":false,"start_date":null,"group_id":"cnzk9zc90jqy","page_id":"pdrzb3d64wsj","group":false,"only_show_if_degraded":false}]}`

		jsBytes := []byte(js)

		sp := statuspage{}
		err := json.Unmarshal(jsBytes, &sp)
		if err != nil {
			log.Printf("Unable to verify HCP status.")
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
			return nil, diag.Errorf("HCP is experiencing an outage. Please check https://status.hashicorp.com/ for more details.")
		case degradedPerformance:
			log.Printf("HCP is experiencing degraded performance. Please check https://status.hashicorp.com/ for more details.")
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "HCP is experiencing degraded performance.",
				Detail:   "Please check https://status.hashicorp.com/ for more details.",
			})
		case underMaintenance:
			log.Printf("HCP is undergoing maintenance that may affect performance. Please check https://status.hashicorp.com/ for more details.")
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "HCP is undergoing maintenance that may affect performance.",
				Detail:   "Please check https://status.hashicorp.com/ for more details.",
			})
		}

		userAgent := p.UserAgent("terraform-provider-hcp", version.ProviderVersion)
		clientID := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)

		client, err := clients.NewClient(clients.ClientConfig{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			SourceChannel: userAgent,
		})
		if err != nil {
			return nil, diag.Errorf("unable to create HCP api client: %v", err)
		}

		// For the initial release, since only one project is allowed per organization, the
		// provider handles fetching the organization's single project, instead of requiring the
		// user to set it. When multiple projects are supported, this helper will be deprecated
		// with a warning: when multiple projects exist within the org, a project ID must be set
		// on the provider or on each resource.
		project, err := getProjectFromCredentials(ctx, client)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		client.Config.OrganizationID = project.Parent.ID
		client.Config.ProjectID = project.ID

		return client, nil
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
