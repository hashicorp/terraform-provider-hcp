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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TemplateResource{}
var _ resource.ResourceWithImportState = &TemplateResource{}

func NewTemplateResource() resource.Resource {
	return &TemplateResource{}
}

// TemplateResource defines the resource implementation.
type TemplateResource struct {
	client *clients.Client
}

// TemplateResourceModel describes the resource data model.
type TemplateResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ProjectID              types.String `tfsdk:"project_id"`
	OrgID                  types.String `tfsdk:"organization_id"`
	Summary                types.String `tfsdk:"summary"`
	Labels                 types.List   `tfsdk:"labels"`
	Description            types.String `tfsdk:"description"`
	ReadmeMarkdownTemplate types.String `tfsdk:"readme_markdown_template"`

	TerraformCloudWorkspace     *tfcWorkspace        `tfsdk:"terraform_cloud_workspace_details"`
	TerraformNoCodeModuleSource types.String         `tfsdk:"terraform_no_code_module_source"`
	TerraformVariableOptions    []*tfcVariableOption `tfsdk:"variable_options"`
}

type tfcWorkspace struct {
	Name types.String `tfsdk:"name"`
	// this refers to the project ID found in Terraform Cloud
	TerraformProjectID types.String `tfsdk:"terraform_project_id"`
}

func (t tfcWorkspace) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                 types.StringType,
		"terraform_project_id": types.StringType,
	}
}

type tfcVariableOption struct {
	Name         types.String `tfsdk:"name"`
	VariableType types.String `tfsdk:"variable_type"`
	Options      types.List   `tfsdk:"options"`
	UserEditable types.Bool   `tfsdk:"user_editable"`
}

func (r *TemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_template"
}

func (r *TemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Waypoint Template resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Template.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Template is located.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Template is located.",
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
				Description: "Instructions for using the template (markdown format supported).",
			},
			"labels": schema.ListAttribute{
				// Computed:    true,
				Optional:    true,
				Description: "List of labels attached to this Template.",
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
						Description: "Name of the Terraform Cloud Project",
					},
					"terraform_project_id": &schema.StringAttribute{
						Required:    true,
						Description: "Terraform Cloud Project ID",
					},
				},
			},
			"terraform_no_code_module_source": schema.StringAttribute{
				Required:    true,
				Description: "Terraform Cloud No-Code Module details",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"variable_options": schema.SetNestedAttribute{
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
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.UniqueValues(),
							},
							Description: "List of options",
						},
						"user_editable": &schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Description: "Whether the variable is editable by the user " +
								"creating an application",
						},
					},
				},
			},
		},
	}
}

func (r *TemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *TemplateResourceModel

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
			UserEditable: v.UserEditable.ValueBool(),
		})
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateApplicationTemplateBody{
		ApplicationTemplate: &waypoint_models.HashicorpCloudWaypointApplicationTemplate{
			Name:         plan.Name.ValueString(),
			Summary:      plan.Summary.ValueString(),
			Labels:       strLabels,
			Description:  plan.Description.ValueString(),
			ModuleSource: plan.TerraformNoCodeModuleSource.ValueString(),
			TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
				Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
				ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
			},
			VariableOptions: varOpts,
		},
	}

	// Decode the base64 encoded readme markdown template to see if it is encoded
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	// If there is an error, we assume that it is because the string is not encoded. This is ok and
	// we will just use the string as is in the ReadmeTemplate field of the model.
	// Eventually the ReadMeMarkdownTemplate field will be deprecated, so the default behavior will be to
	// expect the readme to not be encoded
	if err != nil {
		modelBody.ApplicationTemplate.ReadmeTemplate = plan.ReadmeMarkdownTemplate.ValueString()
	} else {
		modelBody.ApplicationTemplate.ReadmeMarkdownTemplate = readmeBytes
	}

	params := &waypoint_service.WaypointServiceCreateApplicationTemplateParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}
	createTplResp, err := r.client.Waypoint.WaypointServiceCreateApplicationTemplate(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating template", err.Error())
		return
	}

	var appTemplate *waypoint_models.HashicorpCloudWaypointApplicationTemplate
	if createTplResp.Payload != nil {
		appTemplate = createTplResp.Payload.ApplicationTemplate
	}
	if appTemplate == nil {
		resp.Diagnostics.AddError("unknown error creating template", "empty template returned")
		return
	}

	plan.ID = types.StringValue(appTemplate.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(appTemplate.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(appTemplate.Summary)
	plan.TerraformNoCodeModuleSource = types.StringValue(appTemplate.ModuleSource)

	if appTemplate.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace = tfcWorkspace
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

	plan.TerraformVariableOptions, err = readVarOpts(ctx, appTemplate.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created template resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func readVarOpts(
	ctx context.Context,
	v []*waypoint_models.HashicorpCloudWaypointTFModuleVariable,
	d *diag.Diagnostics,
) ([]*tfcVariableOption, error) {
	var varOpts []*tfcVariableOption

	for _, v := range v {
		varWithOpts := &tfcVariableOption{
			Name:         types.StringValue(v.Name),
			VariableType: types.StringValue(v.VariableType),
			UserEditable: types.BoolValue(v.UserEditable),
		}

		optsList, diags := types.ListValueFrom(ctx, types.StringType, v.Options)
		d.Append(diags...)
		if d.HasError() {
			return nil, fmt.Errorf("error reading options for "+
				"variable %q into list of string", v.Name)
		}

		varWithOpts.Options = optsList

		varOpts = append(varOpts, varWithOpts)
	}
	return varOpts, nil
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TemplateResourceModel

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

	client := r.client

	appTemplate, err := clients.GetApplicationTemplateByID(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Template not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Template", err.Error())
		return
	}

	data.ID = types.StringValue(appTemplate.ID)
	data.Name = types.StringValue(appTemplate.Name)
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	data.Summary = types.StringValue(appTemplate.Summary)
	data.TerraformNoCodeModuleSource = types.StringValue(appTemplate.ModuleSource)

	if appTemplate.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(appTemplate.TerraformCloudWorkspaceDetails.ProjectID),
		}
		data.TerraformCloudWorkspace = tfcWorkspace
	}

	data.TerraformVariableOptions, err = readVarOpts(ctx, appTemplate.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
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

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *TemplateResourceModel

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
			UserEditable: v.UserEditable.ValueBool(),
		})
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateApplicationTemplateBody{
		ApplicationTemplate: &waypoint_models.HashicorpCloudWaypointApplicationTemplate{
			Name:         plan.Name.ValueString(),
			Summary:      plan.Summary.ValueString(),
			Labels:       strLabels,
			Description:  plan.Description.ValueString(),
			ModuleSource: plan.TerraformNoCodeModuleSource.ValueString(),
			TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
				Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
				ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
			},
			VariableOptions: varOpts,
		},
	}

	// Decode the base64 encoded readme markdown template to see if it is encoded
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	// If there is an error, we assume that it is because the string is not encoded. This is ok and
	// we will just use the string as is in the ReadmeTemplate field of the model.
	// Eventually the ReadMeMarkdownTemplate field will be deprecated, so the default behavior will be to
	// expect the readme to not be encoded
	if err != nil {
		modelBody.ApplicationTemplate.ReadmeTemplate = plan.ReadmeMarkdownTemplate.ValueString()
	} else {
		modelBody.ApplicationTemplate.ReadmeMarkdownTemplate = readmeBytes
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
		resp.Diagnostics.AddError("unknown error updating template", "empty template returned")
		return
	}

	plan.ID = types.StringValue(appTemplate.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(appTemplate.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(appTemplate.Summary)
	plan.TerraformNoCodeModuleSource = types.StringValue(appTemplate.ModuleSource)

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

	plan.Description = types.StringValue(appTemplate.Description)
	if appTemplate.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdownTemplate = types.StringValue(appTemplate.ReadmeMarkdownTemplate.String())
	if appTemplate.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	plan.TerraformVariableOptions, err = readVarOpts(ctx, appTemplate.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated template resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TemplateResourceModel

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
			tflog.Info(ctx, "Template not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Template",
			err.Error(),
		)
		return
	}
}

func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
