// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &AgentGroupResource{}
var _ resource.ResourceWithImportState = &AgentGroupResource{}

func NewAgentGroupResource() resource.Resource {
	return &AgentGroupResource{}
}

type AgentGroupResource struct {
	client *clients.Client
}

// AgentGroupResourceModel describes the resource data model
type AgentGroupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
}

func (r *AgentGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_agent_group"
}

func (r *AgentGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Agent Group resource manages the lifecycle of an Agent Group.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the Agent Group.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Waypoint project to which the Agent Group belongs.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the Waypoint organization to which the Agent Group belongs.",
				Computed:    true,
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the Agent Group.",
				Optional:    true,
			},
		},
	}
}

func (r *AgentGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AgentGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *AgentGroupResourceModel

	// Read the Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() && !plan.ProjectID.IsNull() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	if !plan.OrgID.IsUnknown() && !plan.OrgID.IsNull() {
		orgID = plan.OrgID.ValueString()
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceCreateAgentGroupBody{
		Group: &waypoint_models.HashicorpCloudWaypointV20241122AgentGroup{},
	}

	modelBody.Group.Name = plan.Name.ValueString()

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		modelBody.Group.Description = plan.Description.ValueString()
	}

	params := &waypoint_service.WaypointServiceCreateAgentGroupParams{
		NamespaceLocationOrganizationID: orgID,
		NamespaceLocationProjectID:      projectID,
		Body:                            modelBody,
	}

	_, err := r.client.Waypoint.WaypointServiceCreateAgentGroup(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Waypoint Agent Group", err.Error())
		return
	}

	// We do not get the agent group back from the API, so we need to get it in a separate request
	getParams := &waypoint_service.WaypointServiceGetAgentGroupParams{
		Name:                            params.Body.Group.Name,
		NamespaceLocationOrganizationID: orgID,
		NamespaceLocationProjectID:      projectID,
	}

	getAgentGroupResp, err := r.client.Waypoint.WaypointServiceGetAgentGroup(getParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Getting Waypoint Agent Group after creation", err.Error())
		return
	}

	var agentGroupModel *waypoint_models.HashicorpCloudWaypointV20241122AgentGroup
	if getAgentGroupResp.Payload != nil {
		agentGroupModel = getAgentGroupResp.Payload.Group
	} else {
		resp.Diagnostics.AddError("Error Getting Waypoint Agent Group after creation", "The response payload was nil.")
		return
	}

	if agentGroupModel == nil {
		resp.Diagnostics.AddError("Error Getting Waypoint Agent Group after creation", "The agent group model was nil.")
		return
	}

	if agentGroupModel.Description != "" {
		plan.Description = types.StringValue(agentGroupModel.Description)
	} else {
		plan.Description = types.StringNull()
	}
	if agentGroupModel.Name != "" {
		plan.Name = types.StringValue(agentGroupModel.Name)
	} else {
		plan.Name = types.StringNull()
	}

	if plan.ProjectID.IsUnknown() {
		plan.ProjectID = types.StringValue(projectID)
	}
	if plan.OrgID.IsUnknown() {
		plan.OrgID = types.StringValue(orgID)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created Agent group resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *AgentGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AgentGroupResourceModel

	// Read the Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() && !data.ProjectID.IsNull() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	if !data.OrgID.IsUnknown() && !data.OrgID.IsNull() {
		orgID = data.OrgID.ValueString()
	}

	client := r.client

	group, err := clients.GetAgentGroup(ctx, client, &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}, data.Name.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// If the group does not exist, remove it from state
			tflog.Info(ctx, "Waypoint Agent Group not found for organization, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Waypoint Agent Group", err.Error())
		return
	}

	if group.Description != "" {
		data.Description = types.StringValue(group.Description)
	} else {
		data.Description = types.StringNull()
	}

	if group.Name != "" {
		data.Name = types.StringValue(group.Name)
	} else {
		data.Name = types.StringNull()
	}

	if data.ProjectID.IsUnknown() {
		data.ProjectID = types.StringValue(projectID)
	}
	if data.OrgID.IsUnknown() {
		data.OrgID = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *AgentGroupResourceModel

	// Read the Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get the current state as well, so we know the current name of the
	// agent group for reference during the update
	var data *AgentGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() && !plan.ProjectID.IsNull() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	if !plan.OrgID.IsUnknown() && !plan.OrgID.IsNull() {
		orgID = plan.OrgID.ValueString()
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceUpdateAgentGroupBody{}

	modelBody.Description = plan.Description.ValueString()

	params := &waypoint_service.WaypointServiceUpdateAgentGroupParams{
		Body:                            modelBody,
		Name:                            data.Name.ValueString(),
		NamespaceLocationOrganizationID: orgID,
		NamespaceLocationProjectID:      projectID,
	}

	agentGroup, err := r.client.Waypoint.WaypointServiceUpdateAgentGroup(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Waypoint Agent Group", err.Error())
		return
	}

	if agentGroup.Payload == nil || agentGroup.Payload.Group == nil {
		resp.Diagnostics.AddError("Error Updating Waypoint Agent Group", "The response payload or group was nil.")
		return
	}

	// Update the plan with the new value
	if agentGroup.Payload.Group.Description != "" {
		plan.Description = types.StringValue(agentGroup.Payload.Group.Description)
	} else {
		plan.Description = types.StringNull()
	}

	if plan.ProjectID.IsUnknown() {
		plan.ProjectID = types.StringValue(projectID)
	}
	if plan.OrgID.IsUnknown() {
		plan.OrgID = types.StringValue(orgID)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Updated Agent group resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AgentGroupResourceModel

	// Read Terraform state data into the model
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

	params := &waypoint_service.WaypointServiceDeleteAgentGroupParams{
		Name:                            data.Name.ValueString(),
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	_, err := r.client.Waypoint.WaypointServiceDeleteAgentGroup(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// If the group does not exist, just return
			tflog.Info(ctx, "Waypoint Agent Group not found for organization, nothing to delete")
			return
		}
		resp.Diagnostics.AddError("Error Deleting Waypoint Agent Group", err.Error())
		return
	}
}

func (r *AgentGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
