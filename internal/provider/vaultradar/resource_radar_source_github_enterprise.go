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
		TypeName:       "_vault_radar_source_github_enterprise",
		SourceType:     "github_enterprise",
		ResourceSchema: githubEnterpriseSourceSchema,
		GetSourceFromPlan: func(ctx context.Context, plan tfsdk.Plan) (radarSource, diag.Diagnostics) {
			var data githubEnterpriseSourceModel
			diags := plan.Get(ctx, &data)
			return &data, diags
		},
		GetSourceFromState: func(ctx context.Context, state tfsdk.State) (radarSource, diag.Diagnostics) {
			var data githubEnterpriseSourceModel
			diags := state.Get(ctx, &data)
			return &data, diags
		}}

}

var githubEnterpriseSourceSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages a GitHub Enterprise Server data source lifecycle in Vault Radar.",
	Attributes: map[string]schema.Attribute{
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
	},
}

type githubEnterpriseSourceModel struct {
	abstractSourceModel
	DomainName         types.String `tfsdk:"domain_name"`
	GitHubOrganization types.String `tfsdk:"github_organization"`
	Token              types.String `tfsdk:"token"`
}

func (d *githubEnterpriseSourceModel) GetName() types.String { return d.GitHubOrganization }

func (d *githubEnterpriseSourceModel) GetConnectionURL() types.String { return d.DomainName }

func (d *githubEnterpriseSourceModel) GetToken() types.String { return d.Token }
