package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceConsulSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul snapshot resource allows users to managed Consul snapshots of an HCP Consul cluster. " +
			"Snapshots currently have a retention policy of 30 days.",
	}
}
