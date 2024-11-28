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

func NewIntegrationSlackSubscriptionResource() resource.Resource {
	return &integrationSubscriptionResource{
		TypeName:           "_vault_radar_integration_slack_subscription",
		SubscriptionSchema: integrationSlackSubscriptionSchema,
		GetSubscriptionFromPlan: func(ctx context.Context, plan tfsdk.Plan) (integrationSubscription, diag.Diagnostics) {
			var sub slackSubscriptionResourceData
			diags := plan.Get(ctx, &sub)
			return &sub, diags
		},
		GetSubscriptionFromState: func(ctx context.Context, state tfsdk.State) (integrationSubscription, diag.Diagnostics) {
			var sub slackSubscriptionResourceData
			diags := state.Get(ctx, &sub)
			return &sub, diags
		},
	}
}

var integrationSlackSubscriptionSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages an Integration Slack Subscription in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Name of subscription. Name must be unique.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"connection_id": schema.StringAttribute{
			Description: "id of the integration slack connection to use for the subscription.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"channel": schema.StringAttribute{
			Description: "Name of the Slack channel that messages should be sent to. Note that HashiCorp Vault Radar will send a test message to verify the channel. Example: dev-ops-team",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
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

type slackSubscriptionResourceData struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ConnectionID types.String `tfsdk:"connection_id"`
	Channel      types.String `tfsdk:"channel"`
	ProjectID    types.String `tfsdk:"project_id"`
}

type slackSubscriptionDetails struct {
	Channel string `json:"channel"`
}

func (d *slackSubscriptionResourceData) GetID() types.String { return d.ID }

func (d *slackSubscriptionResourceData) SetID(id types.String) { d.ID = id }

func (d *slackSubscriptionResourceData) GetProjectID() types.String { return d.ProjectID }

func (d *slackSubscriptionResourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *slackSubscriptionResourceData) GetName() types.String { return d.Name }

func (d *slackSubscriptionResourceData) SetName(name types.String) { d.Name = name }

func (d *slackSubscriptionResourceData) GetConnectionID() types.String { return d.ConnectionID }

func (d *slackSubscriptionResourceData) SetConnectionID(connectionID types.String) {
	d.ConnectionID = connectionID
}

func (d *slackSubscriptionResourceData) GetDetails() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	details := slackSubscriptionDetails{
		Channel: d.Channel.ValueString(),
	}

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error getting Radar Integration Subscription details", err.Error()))
		return "", diags
	}

	return string(detailsBytes), nil
}

func (d *slackSubscriptionResourceData) SetDetails(details string) diag.Diagnostics {
	var diags diag.Diagnostics

	var detailsData slackSubscriptionDetails
	if err := json.Unmarshal([]byte(details), &detailsData); err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error reading Radar Integration Slack Subscription", err.Error()))
		return diags
	}

	d.Channel = types.StringValue(detailsData.Channel)

	return nil
}
