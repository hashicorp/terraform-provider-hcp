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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationTemplateResource{}
var _ resource.ResourceWithImportState = &ApplicationTemplateResource{}

func NewApplicationTemplateResource() resource.Resource {
	return &ApplicationTemplateResource{}
}

// ApplicationTemplateResource defines the resource implementation.
type ApplicationTemplateResource struct {
	client *clients.Client
}

// ApplicationTemplateResourceModel describes the resource data model.
type ApplicationTemplateResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ProjectID              types.String `tfsdk:"project_id"`
	OrgID                  types.String `tfsdk:"organization_id"`
	Summary                types.String `tfsdk:"summary"`
	Labels                 types.List   `tfsdk:"labels"`
	Description            types.String `tfsdk:"description"`
	ReadmeMarkdownTemplate types.String `tfsdk:"readme_markdown_template"`

	TerraformCloudWorkspace  *tfcWorkspace        `tfsdk:"terraform_cloud_workspace_details"`
	TerraformNoCodeModule    *tfcNoCodeModule     `tfsdk:"terraform_no_code_module"`
	TerraformVariableOptions []*tfcVariableOption `tfsdk:"variable_options"`
}

type tfcWorkspace struct {
	Name types.String `tfsdk:"name"`
	// this refers to the project ID found in Terraform Cloud
	TerraformProjectID types.String `tfsdk:"terraform_project_id"`
}

type tfcNoCodeModule struct {
	Source  types.String `tfsdk:"source"`
	Version types.String `tfsdk:"version"`
}

type tfcVariableOption struct {
	Name         types.String `tfsdk:"name"`
	VariableType types.String `tfsdk:"variable_type"`
	Options      types.List   `tfsdk:"options"`
}

func (r *ApplicationTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_application_template"
}

func (r *ApplicationTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Waypoint Application Template resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Application Template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Application Template.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Application Template is located.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Application Template is located.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"summary": schema.StringAttribute{
				Description: "A brief description of the template, up to 110 characters",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the template, along with when and why it should be used, up to 500 characters",
			},
			"readme_markdown_template": schema.StringAttribute{
				Optional:    true,
				Description: "Instructions for using the template (markdown format supported",
			},
			"labels": schema.ListAttribute{
				// Computed:    true,
				Optional:    true,
				Description: "List of labels attached to this Application Template.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"terraform_cloud_workspace_details": &schema.SingleNestedAttribute{
				Required:    true,
				Description: "Terraform Cloud Workspace details",
				Attributes: map[string]schema.Attribute{
					"name": &schema.StringAttribute{
						Required:    true,
						Description: "Name of the Terraform Cloud Workspace",
					},
					"terraform_project_id": &schema.StringAttribute{
						Required:    true,
						Description: "Terraform Cloud Project ID",
					},
				},
			},
			"terraform_no_code_module": &schema.SingleNestedAttribute{
				Required:    true,
				Description: "Terraform Cloud No-Code Module details",
				Attributes: map[string]schema.Attribute{
					"source": &schema.StringAttribute{
						Required:    true,
						Description: "No-Code Module Source",
					},
					"version": &schema.StringAttribute{
						Required:    true,
						Description: "No-Code Module Version",
					},
				},
			},
			"variable_options": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of variable options for the template",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": &schema.StringAttribute{
							Required:    true,
							Description: "Variable name",
						},
						"variable_type": &schema.StringAttribute{
							Required:    true,
							Description: "Variable type",
						},
						"options": &schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "List of options",
						},
					},
				},
			},
		},
	}
}

func (r *ApplicationTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ApplicationTemplateResourceModel

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

	strLabels := []string{}
	diags := plan.Labels.ElementsAs(ctx, &strLabels, false)
	if diags.HasError() {
		return
	}

	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"error decoding the base64 file contents",
			err.Error(),
		)
	}

	varOpts := []*waypoint_models.HashicorpCloudWaypointTFModuleVariable{}
	for _, v := range plan.TerraformVariableOptions {
		strOpts := []string{}
		diags = v.Options.ElementsAs(ctx, &strOpts, false)
		if diags.HasError() {
			return
		}

		varOpts = append(varOpts, &waypoint_models.HashicorpCloudWaypointTFModuleVariable{
			Name:         v.Name.ValueString(),
			VariableType: v.VariableType.ValueString(),
			Options:      strOpts,
			UserEditable: false,
		})
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateApplicationTemplateBody{
		ApplicationTemplate: &waypoint_models.HashicorpCloudWaypointApplicationTemplate{
			Name:                   plan.Name.ValueString(),
			Summary:                plan.Summary.ValueString(),
			Labels:                 strLabels,
			Description:            plan.Description.ValueString(),
			ReadmeMarkdownTemplate: readmeBytes,
			TerraformNocodeModule: &waypoint_models.HashicorpCloudWaypointTerraformNocodeModule{
				Source:  plan.TerraformNoCodeModule.Source.ValueString(),
				Version: plan.TerraformNoCodeModule.Version.ValueString(),
			},
			TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
				Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
				ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
			},
			VariableOptions: varOpts,
		},
	}

	params := &waypoint_service.WaypointServiceCreateApplicationTemplateParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}
	createTplResp, err := r.client.Waypoint.WaypointServiceCreateApplicationTemplate(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application template", err.Error())
		return
	}

	var appTemplate *waypoint_models.HashicorpCloudWaypointApplicationTemplate
	if createTplResp.Payload != nil {
		appTemplate = createTplResp.Payload.ApplicationTemplate
	}
	if appTemplate == nil {
		resp.Diagnostics.AddError("unknown error creating application template", "empty application template returned")
		return
	}

	plan.ID = types.StringValue(appTemplate.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(appTemplate.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(appTemplate.Summary)

	if appTemplate.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace = tfcWorkspace
	}

	if appTemplate.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(appTemplate.TerraformNocodeModule.Source),
			Version: types.StringValue(appTemplate.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule = tfcNoCode
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, appTemplate.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(labels.Elements()) == 0 {
		labels = types.ListNull(types.StringType)
	}

	plan.Labels = labels

	// set plan.description if it's not null or appTemplate.description is not
	// empty
	plan.Description = types.StringValue(appTemplate.Description)
	if appTemplate.Description == "" {
		plan.Description = types.StringNull()
	}
	// set plan.readme if it's not null or appTemplate.readme is not
	// empty
	plan.ReadmeMarkdownTemplate = types.StringValue(appTemplate.ReadmeMarkdownTemplate.String())
	if appTemplate.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	var actualVars []*tfcVariableOption
	for _, v := range appTemplate.VariableOptions {
		varOptsState := &tfcVariableOption{
			Name:         types.StringValue(v.Name),
			VariableType: types.StringValue(v.VariableType),
		}
		varOptsState.Options, diags = types.ListValueFrom(ctx, types.StringType, v.Options)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		actualVars = append(actualVars, varOptsState)
	}
	plan.TerraformVariableOptions = actualVars

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created application template resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ApplicationTemplateResourceModel

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

	appTemplate, err := clients.GetApplicationTemplateByID(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "TFC Config not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading TFC Config", err.Error())
		return
	}

	data.ID = types.StringValue(appTemplate.ID)
	data.Name = types.StringValue(appTemplate.Name)
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	data.Summary = types.StringValue(appTemplate.Summary)

	if appTemplate.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.ProjectID),
		}
		data.TerraformCloudWorkspace = tfcWorkspace
	}

	if appTemplate.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(appTemplate.TerraformNocodeModule.Source),
			Version: types.StringValue(appTemplate.TerraformNocodeModule.Version),
		}
		data.TerraformNoCodeModule = tfcNoCode
	}

	if appTemplate.VariableOptions != nil && len(appTemplate.VariableOptions) > 0 {
		varOpts := []*tfcVariableOption{}
		for _, v := range appTemplate.VariableOptions {
			varOptsState := &tfcVariableOption{
				Name:         types.StringValue(v.Name),
				VariableType: types.StringValue(v.VariableType),
			}

			vOpts, diags := types.ListValueFrom(ctx, types.StringType, v.Options)
			varOptsState.Options = vOpts

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			varOpts = append(varOpts, varOptsState)
		}

		data.TerraformVariableOptions = varOpts
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, appTemplate.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(labels.Elements()) == 0 {
		labels = types.ListNull(types.StringType)
	}
	data.Labels = labels

	// set data.description if it's not null or appTemplate.description is not
	// empty
	data.Description = types.StringValue(appTemplate.Description)
	if appTemplate.Description == "" {
		data.Description = types.StringNull()
	}
	// set data.readme if it's not null or appTemplate.readme is not
	// empty
	data.ReadmeMarkdownTemplate = types.StringValue(appTemplate.ReadmeMarkdownTemplate.String())
	if appTemplate.ReadmeMarkdownTemplate.String() == "" {
		data.ReadmeMarkdownTemplate = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ApplicationTemplateResourceModel

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

	strLabels := []string{}
	diags := plan.Labels.ElementsAs(ctx, &strLabels, false)
	if diags.HasError() {
		return
	}

	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"error decoding the base64 file contents",
			err.Error(),
		)
		return
	}

	varOpts := []*waypoint_models.HashicorpCloudWaypointTFModuleVariable{}
	for _, v := range plan.TerraformVariableOptions {
		strOpts := []string{}
		diags = v.Options.ElementsAs(ctx, &strOpts, false)
		if diags.HasError() {
			return
		}

		varOpts = append(varOpts, &waypoint_models.HashicorpCloudWaypointTFModuleVariable{
			Name:         v.Name.ValueString(),
			VariableType: v.VariableType.ValueString(),
			Options:      strOpts,
			UserEditable: false,
		})
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateApplicationTemplateBody{
		ApplicationTemplate: &waypoint_models.HashicorpCloudWaypointApplicationTemplate{
			Name:                   plan.Name.ValueString(),
			Summary:                plan.Summary.ValueString(),
			Labels:                 strLabels,
			Description:            plan.Description.ValueString(),
			ReadmeMarkdownTemplate: readmeBytes,
			TerraformNocodeModule: &waypoint_models.HashicorpCloudWaypointTerraformNocodeModule{
				Source:  plan.TerraformNoCodeModule.Source.ValueString(),
				Version: plan.TerraformNoCodeModule.Version.ValueString(),
			},
			TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
				Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
				ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
			},
			VariableOptions: varOpts,
		},
	}

	params := &waypoint_service.WaypointServiceUpdateApplicationTemplateParams{
		NamespaceID:                   ns.ID,
		Body:                          modelBody,
		ExistingApplicationTemplateID: plan.ID.ValueString(),
	}
	app, err := r.client.Waypoint.WaypointServiceUpdateApplicationTemplate(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	var appTemplate *waypoint_models.HashicorpCloudWaypointApplicationTemplate
	if app.Payload != nil {
		appTemplate = app.Payload.ApplicationTemplate
	}
	if appTemplate == nil {
		resp.Diagnostics.AddError("unknown error updating application template", "empty application template returned")
		return
	}

	plan.ID = types.StringValue(appTemplate.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(appTemplate.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(appTemplate.Summary)

	labels, diags := types.ListValueFrom(ctx, types.StringType, appTemplate.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	if appTemplate.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace = tfcWorkspace
	}

	if appTemplate.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(appTemplate.TerraformNocodeModule.Source),
			Version: types.StringValue(appTemplate.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule = tfcNoCode
	}

	plan.Description = types.StringValue(appTemplate.Description)
	if appTemplate.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdownTemplate = types.StringValue(appTemplate.ReadmeMarkdownTemplate.String())
	if appTemplate.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	plan.TerraformVariableOptions = []*tfcVariableOption{}
	for _, v := range appTemplate.VariableOptions {
		varOptsState := &tfcVariableOption{
			Name:         types.StringValue(v.Name),
			VariableType: types.StringValue(v.VariableType),
		}
		varOptsState.Options, diags = types.ListValueFrom(ctx, types.StringType, v.Options)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.TerraformVariableOptions = append(plan.TerraformVariableOptions, varOptsState)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated application template resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ApplicationTemplateResourceModel

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
			"Error Deleting TFC Config",
			err.Error(),
		)
		return
	}

	params := &waypoint_service.WaypointServiceDeleteApplicationTemplateParams{
		NamespaceID:           ns.ID,
		ApplicationTemplateID: data.ID.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDeleteApplicationTemplate(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Application Template not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Application Template",
			err.Error(),
		)
		return
	}
}

func (r *ApplicationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
