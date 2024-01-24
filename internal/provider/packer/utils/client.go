package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func CheckClient(client *clients.Client) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if client == nil {
		diags.AddError(
			"Unconfigured HCP Client",
			"Expected a configured HCP Client. Please report this issue to the provider developers.",
		)
	}

	return diags
}
