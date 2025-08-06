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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewSourceGitHubCloudResource() resource.Resource {
	return &radarSourceResource{
		TypeName:       "_vault_radar_source_github_cloud",
		SourceType:     "github_cloud",
		ResourceSchema: githubCloudSourceSchema,
		GetSourceFromPlan: func(ctx context.Context, plan tfsdk.Plan) (radarSource, diag.Diagnostics) {
			var data githubCloudSourceModel
			diags := plan.Get(ctx, &data)
			return &data, diags
		},
		GetSourceFromState: func(ctx context.Context, state tfsdk.State) (radarSource, diag.Diagnostics) {
			var data githubCloudSourceModel
			diags := state.Get(ctx, &data)
			return &data, diags
		}}
}

var githubCloudSourceSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages a GitHub Cloud data source lifecycle in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"github_organization": schema.StringAttribute{
			Description: `GitHub organization Vault Radar will monitor. Example: type "octocat" for the org https://github.com/octocat`,
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

type githubCloudSourceModel struct {
	abstractSourceModel
	GitHubOrganization types.String `tfsdk:"github_organization"`
	Token              types.String `tfsdk:"token"`
}

func (d *githubCloudSourceModel) GetName() types.String { return d.GitHubOrganization }

func (d *githubCloudSourceModel) GetConnectionURL() types.String { return basetypes.NewStringNull() }

func (d *githubCloudSourceModel) GetToken() types.String { return d.Token }
