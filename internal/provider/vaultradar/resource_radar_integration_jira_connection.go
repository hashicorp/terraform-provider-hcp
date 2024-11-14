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
	"github.com/hashicorp/terraform-provider-hcp/internal/hcpvalidator"
)

func NewIntegrationJiraConnectionResource() resource.Resource {
	return &integrationConnectionResource{
		TypeName:         "_vault_radar_integration_jira_connection",
		IntegrationType:  "jira",
		ConnectionSchema: integrationJiraConnectionSchema,
		GetConnectionFromPlan: func(ctx context.Context, plan tfsdk.Plan) (integrationConnection, diag.Diagnostics) {
			var conn jiraConnectionResourceData
			diags := plan.Get(ctx, &conn)
			return &conn, diags
		},
		GetConnectionFromState: func(ctx context.Context, state tfsdk.State) (integrationConnection, diag.Diagnostics) {
			var conn jiraConnectionResourceData
			diags := state.Get(ctx, &conn)
			return &conn, diags
		},
	}
}

var integrationJiraConnectionSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages an Integration Jira Connection in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
		},
		"name": schema.StringAttribute{
			Description: "Name of connection. Name must be unique.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"email": schema.StringAttribute{
			Description: `Jira user's email.`,
			Required:    true,
			Sensitive:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"token": schema.StringAttribute{
			Description: "A Jira API token.",
			Required:    true,
			Sensitive:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"base_url": schema.StringAttribute{
			Description: "The Jira base URL. Example: https://acme.atlassian.net",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				hcpvalidator.URL(),
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

type jiraConnectionResourceData struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Email     types.String `tfsdk:"email"`
	Token     types.String `tfsdk:"token"`
	BaseURL   types.String `tfsdk:"base_url"`
	ProjectID types.String `tfsdk:"project_id"`
}

type jiraAuthKey struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type jiraConnectionDetails struct {
	TenantURL string `json:"tenant_url"`
}

func (d *jiraConnectionResourceData) GetID() types.String { return d.ID }

func (d *jiraConnectionResourceData) SetID(id types.String) { d.ID = id }

func (d *jiraConnectionResourceData) GetProjectID() types.String { return d.ProjectID }

func (d *jiraConnectionResourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *jiraConnectionResourceData) GetName() types.String { return d.Name }

func (d *jiraConnectionResourceData) SetName(name types.String) { d.Name = name }

func (d *jiraConnectionResourceData) GetAuthKey() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	authKey := jiraAuthKey{
		Email: d.Email.ValueString(),
		Token: d.Token.ValueString(),
	}

	authKeyBytes, err := json.Marshal(authKey)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error getting Radar Integration Connection auth key", err.Error()))
		return "", diags
	}

	return string(authKeyBytes), nil
}

func (d *jiraConnectionResourceData) GetDetails() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	details := jiraConnectionDetails{
		TenantURL: d.BaseURL.ValueString(),
	}

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error getting Radar Integration Connection details", err.Error()))
		return "", diags
	}

	return string(detailsBytes), nil
}

func (d *jiraConnectionResourceData) SetDetails(details string) diag.Diagnostics {
	var diags diag.Diagnostics

	var detailsData jiraConnectionDetails
	if err := json.Unmarshal([]byte(details), &detailsData); err != nil {
		diags.Append(diag.NewErrorDiagnostic("Error setting Radar Integration Connection details", err.Error()))
		return diags
	}
	d.BaseURL = types.StringValue(detailsData.TenantURL)

	return nil
}
