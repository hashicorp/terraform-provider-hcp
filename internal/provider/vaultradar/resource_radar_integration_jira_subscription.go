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

func NewIntegrationJiraSubscriptionResource() resource.Resource {
	return &integrationSubscriptionResource{
		TypeName:           "_vault_radar_integration_jira_subscription",
		SubscriptionSchema: integrationJiraSubscriptionSchema,
		GetSubscriptionFromPlan: func(ctx context.Context, plan tfsdk.Plan) (integrationSubscription, diag.Diagnostics) {
			var sub jiraSubscriptionResourceData
			diags := plan.Get(ctx, &sub)
			return &sub, diags
		},
		GetSubscriptionFromState: func(ctx context.Context, state tfsdk.State) (integrationSubscription, diag.Diagnostics) {
			var sub jiraSubscriptionResourceData
			diags := state.Get(ctx, &sub)
			return &sub, diags
		},
	}
}

var integrationJiraSubscriptionSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages an Integration Jira Subscription in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
		},
		"name": schema.StringAttribute{
			Description: "Name of subscription. Name must be unique.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"connection_id": schema.StringAttribute{
			Description: "id of the integration jira connection to use for the subscription.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"jira_project_key": schema.StringAttribute{
			Description: "The name of the project under which the jira issue will be created. Example: OPS",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"issue_type": schema.StringAttribute{
			Description: "The type of issue to be created from the event(s). Example: Task",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},

		// Optional inputs
		"assignee": schema.StringAttribute{
			Description: "The identifier of the Jira user who will be assigned the ticket. In case of Jira Cloud, this will be the Atlassian Account ID of the user. Example: 71509:11bb945b-c0de-4bac-9d57-9f09db2f7bc9",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"message": schema.StringAttribute{
			Description: "This message will be included in the ticket description.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
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

type jiraSubscriptionResourceData struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConnectionID   types.String `tfsdk:"connection_id"`
	JiraProjectKey types.String `tfsdk:"jira_project_key"`
	IssueType      types.String `tfsdk:"issue_type"`
	Assignee       types.String `tfsdk:"assignee"`
	Message        types.String `tfsdk:"message"`
	ProjectID      types.String `tfsdk:"project_id"`
}

type jiraSubscriptionDetails struct {
	ProjectKey   string `json:"project_key"`
	IssueType    string `json:"issuetype"`
	Assignee     string `json:"assignee"`
	Instructions string `json:"instructions"`
}

func (d *jiraSubscriptionResourceData) GetID() types.String { return d.ID }

func (d *jiraSubscriptionResourceData) SetID(id types.String) { d.ID = id }

func (d *jiraSubscriptionResourceData) GetProjectID() types.String { return d.ProjectID }

func (d *jiraSubscriptionResourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *jiraSubscriptionResourceData) GetName() types.String { return d.Name }

func (d *jiraSubscriptionResourceData) SetName(name types.String) { d.Name = name }

func (d *jiraSubscriptionResourceData) GetConnectionID() types.String { return d.ConnectionID }

func (d *jiraSubscriptionResourceData) SetConnectionID(connectionID types.String) {
	d.ConnectionID = connectionID
}

func (d *jiraSubscriptionResourceData) GetDetails() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	details := jiraSubscriptionDetails{
		ProjectKey:   d.JiraProjectKey.ValueString(),
		IssueType:    d.IssueType.ValueString(),
		Assignee:     d.Assignee.ValueString(),
		Instructions: d.Message.ValueString(),
	}

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error getting Radar Integration Subscription details", err.Error()))
		return "", diags
	}

	return string(detailsBytes), nil
}

func (d *jiraSubscriptionResourceData) SetDetails(details string) diag.Diagnostics {
	var diags diag.Diagnostics

	var detailsData jiraSubscriptionDetails
	if err := json.Unmarshal([]byte(details), &detailsData); err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error reading Radar Integration Jira Subscription", err.Error()))
		return diags
	}

	d.JiraProjectKey = types.StringValue(detailsData.ProjectKey)
	d.IssueType = types.StringValue(detailsData.IssueType)

	// Only update the assignee state if the value is not empty.
	if !(d.Assignee.IsNull() && detailsData.Assignee == "") {
		d.Assignee = types.StringValue(detailsData.Assignee)
	}

	// Only update the message state if the value is not empty.
	if !(d.Message.IsNull() && detailsData.Instructions == "") {
		d.Message = types.StringValue(detailsData.Instructions)
	}

	return nil
}
