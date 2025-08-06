// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"

	service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-radar/preview/2023-05-01/client/data_source_registration_service"
)

var (
	_ resource.Resource              = &radarSourceResource{}
	_ resource.ResourceWithConfigure = &radarSourceResource{}
)

var (
	baseSourceSchema = map[string]schema.Attribute{
		"project_id": schema.StringAttribute{
			Description: "The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"detector_type": schema.StringAttribute{
			Optional:    true,
			Description: "The detector type which will monitor this resource. The default is HCP if not specified.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("hcp", "agent"),
			},
		},
	}
)

// radarSourceResource is an implementation for configuring specific types Radar data sources.
// Examples: hcp_vault_radar_source_github_cloud and hcp_vault_radar_source_github_enterprise make use of
// this implementation to define resources with specific schemas, validation, and state details related to their types.
type radarSourceResource struct {
	client             *clients.Client
	TypeName           string
	SourceType         string
	ResourceSchema     schema.Schema
	GetSourceFromPlan  func(ctx context.Context, plan tfsdk.Plan) (radarSource, diag.Diagnostics)
	GetSourceFromState func(ctx context.Context, state tfsdk.State) (radarSource, diag.Diagnostics)
}

func (r *radarSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.TypeName
}

func (r *radarSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: r.ResourceSchema.MarkdownDescription,
		Attributes:          r.ResourceSchema.Attributes,
	}

	for k, v := range baseSourceSchema {
		resp.Schema.Attributes[k] = v
	}
}

// radarSource is the minimal plan/state that a Radar source must have.
type radarSource interface {
	GetProjectID() types.String
	SetProjectID(types.String)
	GetID() types.String
	SetID(types.String)
	GetName() types.String
	GetConnectionURL() types.String
	GetToken() types.String
	GetDetectorType() types.String
}

// base abstraction of Radar datasource model, partially implements radarSource interface
type abstractSourceModel struct {
	ProjectID    types.String `tfsdk:"project_id"`
	ID           types.String `tfsdk:"id"`
	DetectorType types.String `tfsdk:"detector_type"`
}

func (b *abstractSourceModel) GetProjectID() types.String { return b.ProjectID }

func (b *abstractSourceModel) SetProjectID(projectID types.String) { b.ProjectID = projectID }

func (b *abstractSourceModel) GetID() types.String { return b.ID }

func (b *abstractSourceModel) SetID(id types.String) { b.ID = id }

func (b *abstractSourceModel) GetDetectorType() types.String { return b.DetectorType }

func (r *radarSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *radarSourceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *radarSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	src, diags := r.GetSourceFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	body := service.OnboardDataSourceBody{
		Type: r.SourceType,
		Name: src.GetName().ValueString(),

		Token: src.GetToken().ValueString(),
	}

	if !src.GetConnectionURL().IsNull() {
		body.ConnectionURL = src.GetConnectionURL().ValueString()
	}

	if !src.GetDetectorType().IsNull() {
		body.DetectorType = src.GetDetectorType().ValueString()
	}

	res, err := clients.OnboardRadarSource(ctx, r.client, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Radar source", err.Error())
		return
	}

	src.SetID(types.StringValue(res.GetPayload().ID))
	src.SetProjectID(types.StringValue(projectID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &src)...)
}

func (r *radarSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	src, diags := r.GetSourceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	res, err := clients.GetRadarSource(ctx, r.client, projectID, src.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar source not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar source", err.Error())
		return
	}

	// Resource is marked as deleted on the server.
	if res.GetPayload().Deleted {
		// Don't update or remove the state, because its has not been fully deleted server side.
		tflog.Warn(ctx, "Radar source marked for deletion.")
		return
	}

	// The only other state that could change related to this resource is the token, and for obvious reasons we don't
	// return that in the read response. So we don't need to update the state here.
}

func (r *radarSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	src, diags := r.GetSourceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !src.GetProjectID().IsUnknown() {
		projectID = src.GetProjectID().ValueString()
	}

	// Assert resource still exists.
	res, err := clients.GetRadarSource(ctx, r.client, projectID, src.GetID().ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// Resource is no longer on the server.
			tflog.Info(ctx, "Radar source not found, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to get Radar source", err.Error())
		return
	}

	// Resource is already marked as being deleted on the server. Wait for it to be fully deleted.
	if res.GetPayload().Deleted {
		tflog.Info(ctx, "Radar resource already marked for deletion, waiting for full deletion")
		if err := clients.WaitOnOffboardRadarSource(ctx, r.client, projectID, src.GetID().ValueString()); err != nil {
			resp.Diagnostics.AddError("Unable to delete Radar source", err.Error())
			return
		}

		tflog.Trace(ctx, "Deleted Radar resource")
		return
	}

	// Offboard the Radar source.
	if err := clients.OffboardRadarSource(ctx, r.client, projectID, src.GetID().ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete Radar source", err.Error())
		return
	}
}

func (r *radarSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan, planDiags := r.GetSourceFromPlan(ctx, req.Plan)
	resp.Diagnostics.Append(planDiags...)

	state, stateDiags := r.GetSourceFromState(ctx, req.State)
	resp.Diagnostics.Append(stateDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.GetProjectID().IsUnknown() {
		projectID = plan.GetProjectID().ValueString()
	}

	// Check if the token was updated
	if !plan.GetToken().Equal(state.GetToken()) {
		body := service.UpdateDataSourceTokenBody{
			ID:    plan.GetID().ValueString(),
			Token: plan.GetToken().ValueString(),
		}

		if err := clients.UpdateRadarDataSourceToken(ctx, r.client, projectID, body); err != nil {
			resp.Diagnostics.AddError("Error Updating Radar source token", err.Error())
			return
		}
	}

	// Store the updated plan values
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
