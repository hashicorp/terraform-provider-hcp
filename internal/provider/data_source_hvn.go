package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceHvn() *schema.Resource {
	return &schema.Resource{
		Description: "The HVN data source provides information about an existing HashiCorp Virtual Network.",
	}
}
