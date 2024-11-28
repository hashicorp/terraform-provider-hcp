// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewIntegrationSlackConnectionResource() resource.Resource {
	return &integrationConnectionResource{
		TypeName:         "_vault_radar_integration_slack_connection",
		IntegrationType:  "slack",
		ConnectionSchema: integrationSlackConnectionSchema,
		GetConnectionFromPlan: func(ctx context.Context, plan tfsdk.Plan) (integrationConnection, diag.Diagnostics) {
			var conn slackConnectionResourceData
			diags := plan.Get(ctx, &conn)
			return &conn, diags
		},
		GetConnectionFromState: func(ctx context.Context, state tfsdk.State) (integrationConnection, diag.Diagnostics) {
			var conn slackConnectionResourceData
			diags := state.Get(ctx, &conn)
			return &conn, diags
		},
	}
}

var integrationSlackConnectionSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages an Integration Slack Connection in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Name of connection. Name must be unique.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"token": schema.StringAttribute{
			Description: "Slack bot user OAuth token. Example: Bot token strings begin with 'xoxb'.",
			Required:    true,
			Sensitive:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},

		// Optional inputs
		"project_id": schema.StringAttribute{
			Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

type slackConnectionResourceData struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Token     types.String `tfsdk:"token"`
	ProjectID types.String `tfsdk:"project_id"`
}

type slackAuthKey struct {
	Token string `json:"token"`
}

func (d *slackConnectionResourceData) GetID() types.String { return d.ID }

func (d *slackConnectionResourceData) SetID(id types.String) { d.ID = id }

func (d *slackConnectionResourceData) GetProjectID() types.String { return d.ProjectID }

func (d *slackConnectionResourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *slackConnectionResourceData) GetName() types.String { return d.Name }

func (d *slackConnectionResourceData) SetName(name types.String) { d.Name = name }

func (d *slackConnectionResourceData) GetAuthKey() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	authKey := slackAuthKey{
		Token: d.Token.ValueString(),
	}

	authKeyBytes, err := json.Marshal(authKey)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error getting Radar Integration Connection state", err.Error()))
		return "", diags
	}

	return string(authKeyBytes), nil
}

func (d *slackConnectionResourceData) GetDetails() (string, diag.Diagnostics) {
	return "{}", nil
}

func (d *slackConnectionResourceData) SetDetails(string) diag.Diagnostics {
	// no-op
	return nil
}
