// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ActionConfigResource{}
var _ resource.ResourceWithImportState = &ActionConfigResource{}

type ActionConfigResource struct {
	client *clients.Client
}

// ActionConfigModel describes the resource data model.
type ActionConfigResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
	Description types.String `tfsdk:"description"`
	NamespaceID types.String `tfsdk:"namespace_id"`
	ActionURL   types.String `tfsdk:"action_url"`
	// TODO:
	// request
	// created_at
}

func (r *ActionConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_action_config"
}

func (r *ActionConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Action Config resource managed the lifecycle of an Action Config.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Action Config.",
				Computed:    true,
			},
			// TODO: Should Name be optional?
			"name": schema.StringAttribute{
				Description: "The name of the Action Config.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Action Config is located.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Action Config is located.",
				Computed:    true,
			},
			// An Action Config description must be fewer than 125 characters if set.
			"description": schema.StringAttribute{
				Description: "A description of the Action Config.",
				// TODO: Should this be optional?
				Optional: true,
			},
			"namespace_id": schema.StringAttribute{
				Description: "Internal Namespace ID",
				Computed:    true,
			},
			// Action URL is required if the action is custom
			"action_url": schema.StringAttribute{
				Description: "The URL to trigger an action on. Only used in Custom mode",
				Optional:    true,
			},
		},
	}
}

func (r *ActionConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ActionConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ActionConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting namespace by location",
			err.Error(),
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointActionConfig{},
	}

	if !plan.Name.IsUnknown() {
		modelBody.ActionConfig.Name = plan.Name.ValueString()
	}

	if !plan.Description.IsUnknown() {
		modelBody.ActionConfig.Description = plan.Description.ValueString()
	}

	if !plan.ActionURL.IsUnknown() {
		modelBody.ActionConfig.ActionURL = plan.ActionURL.ValueString()
	}

	params := &waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}

	aCfg, err := r.client.Waypoint.WaypointServiceCreateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Action Config", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointActionConfig
	if aCfg.Payload != nil {
		aCfgModel = aCfg.Payload.ActionConfig
	}
	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error creating Action Config", "Empty Action Config returned")
		return
	}

	plan.ID = types.StringValue(aCfgModel.ID)
	plan.Name = types.StringValue(aCfgModel.Name)
	plan.Description = types.StringValue(aCfgModel.Description)
	plan.ActionURL = types.StringValue(aCfgModel.ActionURL)

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.NamespaceID = types.StringValue(ns.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created Action Config resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client

	actionCfg, err := clients.GetActionConfigByID(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action Config not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Action Config", err.Error())
		return
	}

	data.ID = types.StringValue(actionCfg.ID)
	data.Name = types.StringValue(actionCfg.Name)
	data.Description = types.StringValue(actionCfg.Description)
	data.ActionURL = types.StringValue(actionCfg.ActionURL)

	data.ProjectID = types.StringValue(projectID)
	data.OrgID = types.StringValue(orgID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ActionConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get the current state as well, so we know the current name of the
	// action config for reference during the update
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting namespace by location",
			err.Error(),
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointActionConfig{},
	}

	// These are the updated values
	if !plan.Name.IsUnknown() {
		modelBody.ActionConfig.Name = plan.Name.ValueString()
	}
	if !plan.Description.IsUnknown() {
		modelBody.ActionConfig.Description = plan.Description.ValueString()
	}
	if !plan.ActionURL.IsUnknown() {
		modelBody.ActionConfig.ActionURL = plan.ActionURL.ValueString()
	}

	params := &waypoint_service.WaypointServiceUpdateActionConfigParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}

	actionCfg, err := r.client.Waypoint.WaypointServiceUpdateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Action Config", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointActionConfig
	if actionCfg.Payload != nil {
		aCfgModel = actionCfg.Payload.ActionConfig
	}
	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error updating Action Config", "Empty Action Config returned")
		return
	}

	plan.ID = types.StringValue(aCfgModel.ID)
	plan.Name = types.StringValue(aCfgModel.Name)
	plan.Description = types.StringValue(aCfgModel.Description)
	plan.ActionURL = types.StringValue(aCfgModel.ActionURL)

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.NamespaceID = types.StringValue(ns.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Updated Action Config resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ActionConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	client := r.client
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Action Config",
			err.Error(),
		)
		return
	}

	params := &waypoint_service.WaypointServiceDeleteActionConfigParams{
		NamespaceID: ns.ID,
		ActionID:    &data.ID.ValueString(),
		ActionName:  &data.Name.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDeleteActionConfig(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action Config not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Action Config",
			err.Error(),
		)
		return
	}
}
func (r *ActionConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
