package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceVaultCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The Vault cluster resource allows you to manage an HCP Vault cluster.",
	}
}
