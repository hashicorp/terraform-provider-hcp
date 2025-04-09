// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TfcConfigResource{}

func NewTfcConfigResource() resource.Resource {
	return &TfcConfigResource{}
}

// TfcConfigResource defines the resource implementation.
type TfcConfigResource struct {
	client *clients.Client
}

// TfcConfigResourceModel describes the resource data model.
type TfcConfigResourceModel struct {
	// note: there is no true ID in the TFC Config, and each HCP Waypoint
	// organization has only 1 TFC Config, so we use the TFC Organization name
	// as an ID.
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	Token      types.String `tfsdk:"token"`
	TfcOrgName types.String `tfsdk:"tfc_org_name"`
}

func (r *TfcConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_tfc_config"
}

func (r *TfcConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TFC Configuration used by Waypoint to administer TFC workspaces and applications.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Waypoint Project ID to associate with the TFC config",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					// project_id is used in the ID, so if it changes we signal
					// that we must replace this resource. This will force the
					// deletion of the old TFC config for the old project_id,
					// which makes sense because we'll no longer be managing it
					// here.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Terraform Cloud team token. The token must include permissions to manage workspaces and applications.",
			},
			"tfc_org_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Terraform Cloud Organization with which the token is associated.",
				PlanModifiers: []planmodifier.String{
					// tfc_org_name is used in the ID, so if it changes we signal
					// that we must replace this resource. This will force the
					// deletion of the old TFC config.
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *TfcConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TfcConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating TFC Config")
	var plan TfcConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceCreateTFCConfigBody{
		TfcConfig: &waypoint_models.HashicorpCloudWaypointTFCConfig{
			OrganizationName: plan.TfcOrgName.ValueString(),
			Token:            plan.Token.ValueString(),
		},
	}

	params := &waypoint_service.WaypointServiceCreateTFCConfigParams{
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
		Body:                            modelBody,
	}

	config, err := r.client.Waypoint.WaypointServiceCreateTFCConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating TFC Config",
			err.Error(),
		)
		return
	}

	uID := generateUID(projectID)

	plan.ID = types.StringValue(uID)
	plan.TfcOrgName = types.StringValue(config.Payload.TfcConfig.OrganizationName)
	plan.ProjectID = types.StringValue(projectID)

	tflog.Trace(ctx, "Created TFC Config resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TfcConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TfcConfigResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() && !data.ProjectID.IsNull() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	params := &waypoint_service.WaypointServiceGetTFCConfigParams{
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}
	config, err := r.client.Waypoint.WaypointServiceGetTFCConfig(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "TFC Config not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading TFC Config", err.Error())
		return
	}

	if config.Payload == nil || config.Payload.TfcConfig == nil {
		resp.Diagnostics.AddError("Error reading TFC Config", "empty payload")
		return
	}

	uID := generateUID(projectID)

	data.ID = types.StringValue(uID)
	data.TfcOrgName = types.StringValue(config.Payload.TfcConfig.OrganizationName)
	data.ProjectID = types.StringValue(projectID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TfcConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating TFC Config")
	var plan TfcConfigResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceUpdateTFCConfigBody{
		TfcConfig: &waypoint_models.HashicorpCloudWaypointTFCConfig{
			OrganizationName: plan.TfcOrgName.ValueString(),
			Token:            plan.Token.ValueString(),
		},
	}

	params := &waypoint_service.WaypointServiceUpdateTFCConfigParams{
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
		Body:                            modelBody,
	}

	config, err := r.client.Waypoint.WaypointServiceUpdateTFCConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating TFC Config",
			err.Error(),
		)
		return
	}

	uID := generateUID(projectID)

	plan.ID = types.StringValue(uID)
	plan.TfcOrgName = types.StringValue(config.Payload.TfcConfig.OrganizationName)
	plan.ProjectID = types.StringValue(projectID)

	tflog.Trace(ctx, "Updated TFC Config resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TfcConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TfcConfigResourceModel

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

	params := &waypoint_service.WaypointServiceDeleteTFCConfigParams{
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
	}

	_, err := r.client.Waypoint.WaypointServiceDeleteTFCConfig(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "TFC Config not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting TFC Config",
			err.Error(),
		)
		return
	}
}

// getNamespaceByLocation will retrieve a namespace by location information
// provided by HCP
func getNamespaceByLocation(_ context.Context, client *clients.Client, loc *sharedmodels.HashicorpCloudLocationLocation) (*waypoint_models.HashicorpCloudWaypointNamespace, error) {
	// TODO:(clint) consolidate this either in the wrapper or something
	namespaceParams := &waypoint_service.WaypointServiceGetNamespaceParams{
		LocationOrganizationID: loc.OrganizationID,
		LocationProjectID:      loc.ProjectID,
	}
	// get namespace
	ns, err := client.Waypoint.WaypointServiceGetNamespace(namespaceParams, nil)
	if err != nil {
		return nil, err
	}
	return ns.GetPayload().Namespace, nil
}

// Generate the unique ID for the resource
func generateUID(projectID string) string {
	return fmt.Sprintf("/project/%s/%s", projectID, "waypoint_tfc_config")
}
