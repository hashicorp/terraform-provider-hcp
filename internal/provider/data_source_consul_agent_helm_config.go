package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceConsulAgentHelmConfig() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.",
	}
}
