// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"encoding/base64"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource defines the resource implementation.
type ApplicationResource struct {
	client *clients.Client
}

// ApplicationResourceModel describes the resource data model.
type ApplicationResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	ProjectID               types.String `tfsdk:"project_id"`
	OrgID                   types.String `tfsdk:"organization_id"`
	ReadmeMarkdown          types.String `tfsdk:"readme_markdown"`
	ApplicationTemplateID   types.String `tfsdk:"application_template_id"`
	ApplicationTemplateName types.String `tfsdk:"application_template_name"`
	NamespaceID             types.String `tfsdk:"namespace_id"`

	// deferred for now
	// Tags       types.List `tfsdk:"tags"`

	// deferred and probably a list or objects, but may possible be a separate
	// ActionCfgs types.List `tfsdk:"action_cfgs"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Application resource managed the lifecycle of an Application that's based off of a Template.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Application is located.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Application is located.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_template_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Application Template this Application is based on.",
			},
			// application_template_name is a computed only attribute for ease
			// of reference
			"application_template_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Application Template this Application is based on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"readme_markdown": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Description: "Instructions for using the Application (markdown" +
					" format supported). Note: this is a base64 encoded string, and " +
					"can only be set in configuration after initial creation. The" +
					" initial version of the README is generated from the README " +
					"Template from source Application Template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace_id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal Namespace ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ApplicationResourceModel

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

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateApplicationFromTemplateBody{
		Name: plan.Name.ValueString(),
		ApplicationTemplate: &waypoint_models.HashicorpCloudWaypointRefApplicationTemplate{
			ID: plan.ApplicationTemplateID.ValueString(),
		},
	}

	params := &waypoint_service.WaypointServiceCreateApplicationFromTemplateParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}
	app, err := r.client.Waypoint.WaypointServiceCreateApplicationFromTemplate(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application from template", err.Error())
		return
	}

	var application *waypoint_models.HashicorpCloudWaypointApplication
	if app.Payload != nil {
		application = app.Payload.Application
	}
	if application == nil {
		resp.Diagnostics.AddError("unknown error creating application from template", "empty application returned")
		return
	}

	plan.ID = types.StringValue(application.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(application.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.ApplicationTemplateName = types.StringValue(application.ApplicationTemplate.Name)
	plan.NamespaceID = types.StringValue(ns.ID)

	// set plan.readme if it's not null or application.readme is not
	// empty
	plan.ReadmeMarkdown = types.StringValue(application.ReadmeMarkdown.String())
	if application.ReadmeMarkdown.String() == "" {
		plan.ReadmeMarkdown = types.StringNull()
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created application from template resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ApplicationResourceModel

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

	application, err := clients.GetApplicationByID(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Application not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Application", err.Error())
		return
	}

	data.ID = types.StringValue(application.ID)
	data.ProjectID = types.StringValue(projectID)
	data.Name = types.StringValue(application.Name)
	data.OrgID = types.StringValue(orgID)
	data.ApplicationTemplateName = types.StringValue(application.ApplicationTemplate.Name)

	// set plan.readme if it's not null or application.readme is not
	// empty
	data.ReadmeMarkdown = types.StringValue(application.ReadmeMarkdown.String())
	if application.ReadmeMarkdown.String() == "" {
		data.ReadmeMarkdown = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ApplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get the current state as well, so we know the current name of the
	// application for reference during the update
	var data *ApplicationResourceModel

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

	// read the readme from the plan and decode it
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdown.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"error decoding the base64 file contents",
			err.Error(),
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateApplicationBody{
		// this is the updated name
		Name:           plan.Name.ValueString(),
		ReadmeMarkdown: readmeBytes,
	}

	params := &waypoint_service.WaypointServiceUpdateApplicationParams{
		ApplicationID: plan.ID.ValueString(),
		NamespaceID:   ns.ID,
		Body:          modelBody,
	}
	app, err := r.client.Waypoint.WaypointServiceUpdateApplication(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Application", err.Error())
		return
	}

	var application *waypoint_models.HashicorpCloudWaypointApplication
	if app.Payload != nil {
		application = app.Payload.Application
	}
	if application == nil {
		resp.Diagnostics.AddError("unknown error updating application", "empty application returned")
		return
	}

	plan.ID = types.StringValue(application.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(application.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.ApplicationTemplateName = types.StringValue(application.ApplicationTemplate.Name)
	plan.NamespaceID = types.StringValue(ns.ID)

	// set plan.readme if it's not null or application.readme is not
	// empty
	plan.ReadmeMarkdown = types.StringValue(application.ReadmeMarkdown.String())
	if application.ReadmeMarkdown.String() == "" {
		plan.ReadmeMarkdown = types.StringNull()
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated application resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ApplicationResourceModel

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
			"error deleting application",
			err.Error(),
		)
		return
	}

	params := &waypoint_service.WaypointServiceDestroyApplicationParams{
		NamespaceID:   ns.ID,
		ApplicationID: data.ID.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDestroyApplication(params, nil)

	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Application not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"error deleting Application",
			err.Error(),
		)
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
