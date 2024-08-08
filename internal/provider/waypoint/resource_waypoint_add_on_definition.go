// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"encoding/base64"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypointModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AddOnDefinitionResource{}
var _ resource.ResourceWithImportState = &AddOnDefinitionResource{}

func NewAddOnDefinitionResource() resource.Resource {
	return &AddOnDefinitionResource{}
}

// AddOnDefinitionResource defines the resource implementation.
type AddOnDefinitionResource struct {
	client *clients.Client
}

// AddOnDefinitionResourceModel describes the resource data model.
type AddOnDefinitionResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ProjectID              types.String `tfsdk:"project_id"`
	OrgID                  types.String `tfsdk:"organization_id"`
	Summary                types.String `tfsdk:"summary"`
	Labels                 types.List   `tfsdk:"labels"`
	Description            types.String `tfsdk:"description"`
	ReadmeMarkdownTemplate types.String `tfsdk:"readme_markdown_template"`

	TerraformCloudWorkspace     types.Object         `tfsdk:"terraform_cloud_workspace_details"`
	TerraformNoCodeModuleSource types.String         `tfsdk:"terraform_no_code_module_source"`
	TerraformVariableOptions    []*tfcVariableOption `tfsdk:"variable_options"`
}

func (r *AddOnDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_add_on_definition"
}

func (r *AddOnDefinitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Waypoint Add-on Definition resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Add-on Definition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Add-on Definition.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Add-on Definition is located.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Add-on Definition is located.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"summary": schema.StringAttribute{
				Description: "A short summary of the Add-on Definition.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A longer description of the Add-on Definition.",
				Required:    true,
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				Description: "List of labels attached to this Add-on Definition.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"readme_markdown_template": schema.StringAttribute{
				Description: "The markdown template for the Add-on Definition README (markdown format supported).",
				Optional:    true,
			},
			"terraform_cloud_workspace_details": &schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Description: "Terraform Cloud Workspace details. If not provided, defaults to " +
					"the HCP Terraform project of the associated application.",
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
			"terraform_no_code_module_source": &schema.StringAttribute{
				Required: true,
				Description: "Terraform Cloud no-code Module Source, expected to be in one of the following formats:" +
					" \"app.terraform.io/hcp_waypoint_example/ecs-advanced-microservice/aws\" or " +
					"\"private/hcp_waypoint_example/ecs-advanced-microservice/aws\"",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"variable_options": schema.SetNestedAttribute{
				Optional:    true,
				Description: "List of variable options for the Add-on Definition.",
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
						"user_editable": &schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Description: "Whether the variable is editable by the user creating an " +
								"add-on. If options are provided, then the user may only use those " +
								"options, regardless of this setting.",
						},
					},
				},
			},
		},
	}
}

func (r *AddOnDefinitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AddOnDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *AddOnDefinitionResourceModel

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

	stringLabels := []string{}
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		diagnostics := plan.Labels.ElementsAs(ctx, &stringLabels, false)
		if diagnostics.HasError() {
			resp.Diagnostics.AddError(
				"error converting labels",
				"The list of labels was incorrectly formated",
			)
			return
		}
	}

	varOpts := []*waypointModels.HashicorpCloudWaypointTFModuleVariable{}
	for _, v := range plan.TerraformVariableOptions {
		strOpts := []string{}
		diags := v.Options.ElementsAs(ctx, &strOpts, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		varOpts = append(varOpts, &waypointModels.HashicorpCloudWaypointTFModuleVariable{
			Name:         v.Name.ValueString(),
			VariableType: v.VariableType.ValueString(),
			Options:      strOpts,
			UserEditable: v.UserEditable.ValueBool(),
		})
	}

	modelBody := &waypointModels.HashicorpCloudWaypointWaypointServiceCreateAddOnDefinitionBody{
		Name:            plan.Name.ValueString(),
		Summary:         plan.Summary.ValueString(),
		Description:     plan.Description.ValueString(),
		Labels:          stringLabels,
		ModuleSource:    plan.TerraformNoCodeModuleSource.ValueString(),
		VariableOptions: varOpts,
	}

	// Decode the base64 encoded readme markdown template to see if it is encoded
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	// If there is an error, we assume that it is because the string is not encoded. This is ok and
	// we will just use the string as is in the ReadmeTemplate field of the model.
	// Eventually the ReadMeMarkdownTemplate field will be deprecated, so the default behavior will be to
	// expect the readme to not be encoded
	if err != nil {
		modelBody.ReadmeTemplate = plan.ReadmeMarkdownTemplate.ValueString()
	} else {
		modelBody.ReadmeMarkdownTemplate = readmeBytes
	}

	if !plan.TerraformCloudWorkspace.IsNull() && !plan.TerraformCloudWorkspace.IsUnknown() {
		workspaceDetails := &tfcWorkspace{}
		diags := plan.TerraformCloudWorkspace.As(ctx, workspaceDetails, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		modelBody.TerraformCloudWorkspaceDetails = &waypointModels.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      workspaceDetails.Name.ValueString(),
			ProjectID: workspaceDetails.TerraformProjectID.ValueString(),
		}
	}

	params := &waypoint_service.WaypointServiceCreateAddOnDefinitionParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}
	def, err := r.client.Waypoint.WaypointServiceCreateAddOnDefinition(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating add-on definition", err.Error())
		return
	}

	var addOnDefinition *waypointModels.HashicorpCloudWaypointAddOnDefinition
	if def.Payload != nil {
		addOnDefinition = def.Payload.AddOnDefinition
	}
	if addOnDefinition == nil {
		resp.Diagnostics.AddError("unknown error creating add-on definition", "empty add-on definition returned")
		return
	}

	plan.ID = types.StringValue(addOnDefinition.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(addOnDefinition.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOnDefinition.Summary)
	plan.TerraformNoCodeModuleSource = types.StringValue(addOnDefinition.ModuleSource)

	plan.Description = types.StringValue(addOnDefinition.Description)
	// set plan.description if it's not null or addOnDefinition.description is not empty
	if addOnDefinition.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdownTemplate = types.StringValue(addOnDefinition.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOnDefinition.readme is not empty
	if addOnDefinition.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOnDefinition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	if addOnDefinition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace, diags = types.ObjectValueFrom(ctx, tfcWorkspace.attrTypes(), tfcWorkspace)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	plan.TerraformVariableOptions, err = readVarOpts(ctx, addOnDefinition.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created add-on definition resource")

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AddOnDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *AddOnDefinitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() && !state.ProjectID.IsNull() {
		projectID = state.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client

	definition, err := clients.GetAddOnDefinitionByID(ctx, client, loc, state.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Add-on Definition not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Add-on Definition", err.Error())
		return
	}

	state.ID = types.StringValue(definition.ID)
	state.Name = types.StringValue(definition.Name)
	state.OrgID = types.StringValue(client.Config.OrganizationID)
	state.ProjectID = types.StringValue(client.Config.ProjectID)
	state.Summary = types.StringValue(definition.Summary)
	state.TerraformNoCodeModuleSource = types.StringValue(definition.ModuleSource)

	state.Description = types.StringValue(definition.Description)
	// set plan.description if it's not null or addOnDefinition.description is not empty
	if definition.Description == "" {
		state.Description = types.StringNull()
	}
	state.ReadmeMarkdownTemplate = types.StringValue(definition.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOnDefinition.readme is not empty
	if definition.ReadmeMarkdownTemplate.String() == "" {
		state.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, definition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Labels = labels

	if definition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(definition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(definition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		state.TerraformCloudWorkspace, diags = types.ObjectValueFrom(ctx, tfcWorkspace.attrTypes(), tfcWorkspace)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	state.TerraformVariableOptions, err = readVarOpts(ctx, definition.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AddOnDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *AddOnDefinitionResourceModel

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

	stringLabels := []string{}
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		diagnostics := plan.Labels.ElementsAs(ctx, &stringLabels, false)
		if diagnostics.HasError() {
			resp.Diagnostics.AddError(
				"error converting labels",
				"Failed to convert labels from types.List to string list",
			)
			return
		}
	}

	varOpts := []*waypointModels.HashicorpCloudWaypointTFModuleVariable{}
	for _, v := range plan.TerraformVariableOptions {
		strOpts := []string{}
		diags := v.Options.ElementsAs(ctx, &strOpts, false)
		if diags.HasError() {
			return
		}

		varOpts = append(varOpts, &waypointModels.HashicorpCloudWaypointTFModuleVariable{
			Name:         v.Name.ValueString(),
			VariableType: v.VariableType.ValueString(),
			Options:      strOpts,
			UserEditable: v.UserEditable.ValueBool(),
		})
	}

	// TODO: add support for Tags
	modelBody := &waypointModels.HashicorpCloudWaypointWaypointServiceUpdateAddOnDefinitionBody{
		Name:            plan.Name.ValueString(),
		Summary:         plan.Summary.ValueString(),
		Description:     plan.Description.ValueString(),
		Labels:          stringLabels,
		ModuleSource:    plan.TerraformNoCodeModuleSource.ValueString(),
		VariableOptions: varOpts,
	}

	// Decode the base64 encoded readme markdown template to see if it is encoded
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	// If there is an error, we assume that it is because the string is not encoded. This is ok and
	// we will just use the string as is in the ReadmeTemplate field of the model.
	// Eventually the ReadMeMarkdownTemplate field will be deprecated, so the default behavior will be to
	// expect the readme to not be encoded
	if err != nil {
		modelBody.ReadmeTemplate = plan.ReadmeMarkdownTemplate.ValueString()
	} else {
		modelBody.ReadmeMarkdownTemplate = readmeBytes
	}

	if !plan.TerraformCloudWorkspace.IsNull() && !plan.TerraformCloudWorkspace.IsUnknown() {
		workspaceDetails := &tfcWorkspace{}
		diags := plan.TerraformCloudWorkspace.As(ctx, workspaceDetails, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		modelBody.TerraformCloudWorkspaceDetails = &waypointModels.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      workspaceDetails.Name.ValueString(),
			ProjectID: workspaceDetails.TerraformProjectID.ValueString(),
		}
	}

	params := &waypoint_service.WaypointServiceUpdateAddOnDefinitionParams{
		NamespaceID:               ns.ID,
		Body:                      modelBody,
		ExistingAddOnDefinitionID: plan.ID.ValueString(),
	}
	def, err := r.client.Waypoint.WaypointServiceUpdateAddOnDefinition(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Add-on Definition", err.Error())
		return
	}

	var addOnDefinition *waypointModels.HashicorpCloudWaypointAddOnDefinition
	if def.Payload != nil {
		addOnDefinition = def.Payload.AddOnDefinition
	}
	if addOnDefinition == nil {
		resp.Diagnostics.AddError("Unknown error updating Add-on Definition", "Empty Add-on Definition found")
		return
	}

	plan.ID = types.StringValue(addOnDefinition.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(addOnDefinition.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOnDefinition.Summary)
	plan.TerraformNoCodeModuleSource = types.StringValue(addOnDefinition.ModuleSource)

	plan.Description = types.StringValue(addOnDefinition.Description)
	// set plan.description if it's not null or addOnDefinition.description is not empty
	if addOnDefinition.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdownTemplate = types.StringValue(addOnDefinition.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOnDefinition.readme is not empty
	if addOnDefinition.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOnDefinition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	if addOnDefinition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace, diags = types.ObjectValueFrom(ctx, tfcWorkspace.attrTypes(), tfcWorkspace)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	plan.TerraformVariableOptions, err = readVarOpts(ctx, addOnDefinition.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated add-on definition resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AddOnDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *AddOnDefinitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
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

	params := &waypoint_service.WaypointServiceDeleteAddOnDefinitionParams{
		NamespaceID:       ns.ID,
		AddOnDefinitionID: state.ID.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDeleteAddOnDefinition(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Add-on Definition not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Add-on Definition",
			err.Error(),
		)
		return
	}
}

func (r *AddOnDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
