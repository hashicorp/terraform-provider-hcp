package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceConsulAgentKubernetesSecret() *schema.Resource {
	return &schema.Resource{
		Description: "The agent config Kubernetes secret data source provides Consul agents running in Kubernetes the configuration needed to connect to the Consul cluster.",
	}
}
