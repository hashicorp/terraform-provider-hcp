// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"regexp"

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

func NewSourceGitHubEnterpriseResource() resource.Resource {
	return &radarSourceResource{
		TypeName:         "_vault_radar_source_github_enterprise",
		SourceType:       "github_enterprise",
		ConnectionSchema: githubEnterpriseSourceSchema,
		GetSourceFromPlan: func(ctx context.Context, plan tfsdk.Plan) (radarSource, diag.Diagnostics) {
			var data githubEnterpriseSourceData
			diags := plan.Get(ctx, &data)
			return &data, diags
		},
		GetSourceFromState: func(ctx context.Context, state tfsdk.State) (radarSource, diag.Diagnostics) {
			var data githubEnterpriseSourceData
			diags := state.Get(ctx, &data)
			return &data, diags
		}}

}

var githubEnterpriseSourceSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages a GitHub Enterprise Server data source lifecycle in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"domain_name": schema.StringAttribute{
			Description: "Fully qualified domain name of the server. (Example: myserver.acme.com)",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^(?:[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$`),
					"must be a valid domain name",
				),
			},
		},
		"github_organization": schema.StringAttribute{
			Description: `GitHub organization Vault Radar will monitor. Example: "octocat" for the org https://yourcodeserver.com/octocat`,
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`),
					"must contain only letters, numbers, hyphens, underscores, or periods",
				),
			},
		},
		"token": schema.StringAttribute{
			Description: "GitHub personal access token.",
			Required:    true,
			Sensitive:   true,
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

type githubEnterpriseSourceData struct {
	ID                 types.String `tfsdk:"id"`
	DomainName         types.String `tfsdk:"domain_name"`
	GitHubOrganization types.String `tfsdk:"github_organization"`
	Token              types.String `tfsdk:"token"`
	ProjectID          types.String `tfsdk:"project_id"`
}

func (d *githubEnterpriseSourceData) GetProjectID() types.String { return d.ProjectID }

func (d *githubEnterpriseSourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *githubEnterpriseSourceData) GetID() types.String { return d.ID }

func (d *githubEnterpriseSourceData) SetID(id types.String) { d.ID = id }

func (d *githubEnterpriseSourceData) GetName() types.String { return d.GitHubOrganization }

func (d *githubEnterpriseSourceData) GetConnectionURL() types.String { return d.DomainName }

func (d *githubEnterpriseSourceData) GetToken() types.String { return d.Token }
