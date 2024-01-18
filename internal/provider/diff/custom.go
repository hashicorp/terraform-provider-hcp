// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttrGettable is a small enabler for helper functions that need to read one
// attribute of a Configuration, Plan, or State.
type AttrGettable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// ModifyPlanForDefaultProjectChange modifies a resource plan if the provider's default project has changed and differs from
// the resource's project when it doesn't set the project_id itself.
// Modifying the resource plan makes the resource sensitive to the provider's default configuration updates keeping the
// resource in sync with the configuration inherited from the provider.
func ModifyPlanForDefaultProjectChange(ctx context.Context, providerDefaultProject string, state tfsdk.State, configAttributes, planAttributes AttrGettable, resp *resource.ModifyPlanResponse) {
	if state.Raw.IsNull() {
		return
	}

	orgPath := path.Root("project_id")

	var configProject, plannedProject types.String
	resp.Diagnostics.Append(configAttributes.GetAttribute(ctx, orgPath, &configProject)...)
	resp.Diagnostics.Append(planAttributes.GetAttribute(ctx, orgPath, &plannedProject)...)

	if configProject.IsNull() && !plannedProject.IsNull() && providerDefaultProject != plannedProject.ValueString() {
		// There is no project configured on the resource, yet the provider project is different from
		// the planned project value. We must conclude that the provider default project changed.
		resp.Plan.SetAttribute(ctx, orgPath, types.StringValue(providerDefaultProject))
		resp.RequiresReplace.Append(orgPath)
	}
}
