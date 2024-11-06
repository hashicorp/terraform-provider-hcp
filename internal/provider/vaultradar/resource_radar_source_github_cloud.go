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
		TypeName:         "_vault_radar_source_github_cloud",
		SourceType:       "github_cloud",
		ConnectionSchema: githubCloudSourceSchema,
		GetSourceFromPlan: func(ctx context.Context, plan tfsdk.Plan) (radarSource, diag.Diagnostics) {
			var data githubCloudSourceData
			diags := plan.Get(ctx, &data)
			return &data, diags
		},
		GetSourceFromState: func(ctx context.Context, state tfsdk.State) (radarSource, diag.Diagnostics) {
			var data githubCloudSourceData
			diags := state.Get(ctx, &data)
			return &data, diags
		}}
}

var githubCloudSourceSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages a GitHub Cloud data source lifecycle in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
		},
		"github_organization": schema.StringAttribute{
			Description: `GitHub organization Vault Radar will monitor. Example: type "octocat" for the org https://github.com/octocat`,
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
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

type githubCloudSourceData struct {
	ID                 types.String `tfsdk:"id"`
	GitHubOrganization types.String `tfsdk:"github_organization"`
	Token              types.String `tfsdk:"token"`
	ProjectID          types.String `tfsdk:"project_id"`
}

func (d *githubCloudSourceData) GetProjectID() types.String { return d.ProjectID }

func (d *githubCloudSourceData) SetProjectID(projectID types.String) { d.ProjectID = projectID }

func (d *githubCloudSourceData) GetID() types.String { return d.ID }

func (d *githubCloudSourceData) SetID(id types.String) { d.ID = id }

func (d *githubCloudSourceData) GetName() types.String { return d.GitHubOrganization }

func (d *githubCloudSourceData) GetConnectionURL() types.String { return basetypes.NewStringNull() }

func (d *githubCloudSourceData) GetToken() types.String { return d.Token }
