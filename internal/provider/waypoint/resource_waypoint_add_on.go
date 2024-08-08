// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"
	"strconv"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
var _ resource.Resource = &AddOnResource{}
var _ resource.ResourceWithImportState = &AddOnResource{}

func NewAddOnResource() resource.Resource {
	return &AddOnResource{}
}

// AddOnResource defines the resource implementation.
type AddOnResource struct {
	client *clients.Client
}

// AddOnResourceModel describes the resource data model.
type AddOnResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrgID          types.String `tfsdk:"organization_id"`
	Summary        types.String `tfsdk:"summary"`
	Labels         types.List   `tfsdk:"labels"`
	Description    types.String `tfsdk:"description"`
	ReadmeMarkdown types.String `tfsdk:"readme_markdown"`
	CreatedBy      types.String `tfsdk:"created_by"`
	Count          types.Int64  `tfsdk:"install_count"`
	Status         types.Int64  `tfsdk:"status"`
	ApplicationID  types.String `tfsdk:"application_id"`
	DefinitionID   types.String `tfsdk:"definition_id"`
	OutputValues   types.List   `tfsdk:"output_values"`

	TerraformNoCodeModuleSource types.String `tfsdk:"terraform_no_code_module_source"`

	InputVars                types.Set `tfsdk:"add_on_input_variables"`
	AddOnDefinitionInputVars types.Set `tfsdk:"add_on_definition_input_variables"`
}

type outputValue struct {
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Value     types.String `tfsdk:"value"`
	Sensitive types.Bool   `tfsdk:"sensitive"`
}

func (o outputValue) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":      types.StringType,
		"type":      types.StringType,
		"value":     types.StringType,
		"sensitive": types.BoolType,
	}
}

func (r *AddOnResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_add_on"
}

// TODO: Make most of these computed because they are not used in the protos (Also add variables later)
func (r *AddOnResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Waypoint Add-on resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Add-on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Add-on.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint AddOn is located.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint AddOn is located.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"summary": schema.StringAttribute{
				Description: "A short summary of the Add-on.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A longer description of the Add-on.",
				Computed:    true,
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Description: "List of labels attached to this Add-on.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"readme_markdown": schema.StringAttribute{
				Description: "The markdown for the Add-on README.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The user who created the Add-on.",
				Computed:    true,
			},
			"install_count": schema.Int64Attribute{
				Description: "The number of installed Add-ons for the same Application that share the same " +
					"Add-on Definition.",
				Computed: true,
			},
			"terraform_no_code_module_source": schema.StringAttribute{
				Description: "The Terraform No Code Module source for the Add-on.",
				Computed:    true,
			},
			"definition_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Add-on Definition that this Add-on is created from.",
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Application that this Add-on is created for.",
			},
			"status": schema.Int64Attribute{
				Computed:    true,
				Description: "The status of the Terraform run for the Add-on.",
			},
			"output_values": schema.ListNestedAttribute{
				Computed: true,
				Description: "The output values, stored by HCP Waypoint, of the Terraform run for the Add-on, sensitive values have type " +
					"and value omitted.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the output value.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the output value.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the output value.",
						},
						"sensitive": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the output value is sensitive.",
						},
					},
				},
			},
			"add_on_input_variables": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Input variables set for the add-on.",
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
						"value": &schema.StringAttribute{
							Required:    true,
							Description: "Variable value",
						},
					},
				},
			},
			"add_on_definition_input_variables": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Input variables set for the add-on definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": &schema.StringAttribute{
							Required:    true,
							Description: "Variable name",
						},
						"variable_type": &schema.StringAttribute{
							Optional:    true,
							Description: "Variable type",
						},
						"value": &schema.StringAttribute{
							Required:    true,
							Description: "Variable value",
						},
					},
				},
			},
		},
	}
}

func (r *AddOnResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// TODO: Add support for new fields
func (r *AddOnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *AddOnResourceModel

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

	// varTypes is used to store the variable type for each input variable
	// to be used later when fetching the input variables from the API
	varTypes := map[string]string{}

	// Prepare the input variables that the user provided to the application
	// creation request
	ivs := make([]*waypoint_models.HashicorpCloudWaypointInputVariable, 0)

	var inputVarsSlice []InputVar
	diags := plan.InputVars.ElementsAs(ctx, &inputVarsSlice, false)
	if diags.HasError() {
		return
	}
	for _, v := range inputVarsSlice {
		// add the input variable to the list of input variables for the app
		// creation API call
		ivs = append(ivs, &waypoint_models.HashicorpCloudWaypointInputVariable{
			Name:         v.Name.ValueString(),
			Value:        v.Value.ValueString(),
			VariableType: v.VariableType.ValueString(),
		})

		// store var type for later use when fetching the input variables from the API
		varTypes[v.Name.ValueString()] = v.VariableType.ValueString()
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

	// An application ref can only have one of ID or Name set,
	// we ask for ID, so we will set ID
	applicationID := plan.ApplicationID.ValueString()
	applicationRefModel := &waypoint_models.HashicorpCloudWaypointRefApplication{}
	if applicationID != "" {
		applicationRefModel.ID = applicationID
	} else {
		resp.Diagnostics.AddError(
			"error reading application ID",
			"The application ID was missing",
		)
		return
	}

	// Similarly, a definition ref can only have one of ID or Name set,
	// we ask for ID, so we will set ID
	definitionID := plan.DefinitionID.ValueString()
	definitionRefModel := &waypoint_models.HashicorpCloudWaypointRefAddOnDefinition{}
	if definitionID != "" {
		definitionRefModel.ID = definitionID
	} else {
		resp.Diagnostics.AddError(
			"error reading definition ID",
			"The definition ID was missing",
		)
		return
	}

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateAddOnBody{
		Name:        plan.Name.ValueString(),
		Application: applicationRefModel,
		Definition:  definitionRefModel,
		Variables:   ivs,
	}

	params := &waypoint_service.WaypointServiceCreateAddOnParams{
		NamespaceID: ns.ID,
		Body:        modelBody,
	}
	responseAddOn, err := r.client.Waypoint.WaypointServiceCreateAddOn(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating add-on", err.Error())
		return
	}

	var addOn *waypoint_models.HashicorpCloudWaypointAddOn
	if responseAddOn.Payload != nil {
		addOn = responseAddOn.Payload.AddOn
	}
	if addOn == nil {
		resp.Diagnostics.AddError("unknown error creating add-on", "empty add-on returned")
		return
	}

	plan.ID = types.StringValue(addOn.ID)
	plan.Name = types.StringValue(addOn.Name)
	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOn.Summary)
	plan.Description = types.StringValue(addOn.Description)
	plan.TerraformNoCodeModuleSource = types.StringValue(addOn.TerraformNocodeModule.Source)

	// set plan.description if it's not null or addOn.description is not empty
	if addOn.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdown = types.StringValue(addOn.ReadmeMarkdown.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdown.String() == "" {
		plan.ReadmeMarkdown = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	// Display the reference to the Definition in the plan
	if addOn.Definition != nil {
		if addOn.Definition.ID != "" {
			plan.DefinitionID = types.StringValue(addOn.Definition.ID)
		}
	}

	// Display the reference to the Application in the plan
	if addOn.Application != nil {
		if addOn.Application.ID != "" {
			plan.ApplicationID = types.StringValue(addOn.Application.ID)
		}
	}

	plan.CreatedBy = types.StringValue(addOn.CreatedBy)

	// If we can process status as an int64, add it to the plan
	statusNum, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-on status", err.Error())
	} else {
		plan.Status = types.Int64Value(statusNum)
	}

	// If we can process count as an int64, add it to the plan
	installedCount, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-ons count", err.Error())
	} else {
		plan.Count = types.Int64Value(installedCount)
	}

	ol := readOutputs(addOn.OutputValues)
	if len(ol) > 0 {
		plan.OutputValues, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: outputValue{}.attrTypes()}, ol)
	} else {
		plan.OutputValues = types.ListNull(types.ObjectType{AttrTypes: outputValue{}.attrTypes()})
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputVars, err := clients.GetInputVariables(ctx, client, plan.Name.ValueString(), loc)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to fetch add-on's input variables.")
		return
	}

	addOnInputVars, templateInputVars := splitInputs(inputVars, varTypes)
	if len(addOnInputVars) > 0 {
		aivs, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: InputVar{}.attrTypes()}, addOnInputVars)
		if diags.HasError() {
			return
		}
		plan.InputVars = aivs
	} else {
		plan.InputVars = types.SetNull(types.ObjectType{AttrTypes: InputVar{}.attrTypes()})
	}

	if len(templateInputVars) > 0 {
		tivs, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: InputVar{}.attrTypes()}, templateInputVars)
		if diags.HasError() {
			return
		}
		plan.AddOnDefinitionInputVars = tivs
	} else {
		plan.AddOnDefinitionInputVars = types.SetNull(types.ObjectType{AttrTypes: InputVar{}.attrTypes()})
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created add-on resource")

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// TODO: Add support for new fields
func (r *AddOnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *AddOnResourceModel

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

	// varTypes is used to store the variable type for each input variable
	// to be used later when fetching the input variables from the API
	varTypes := map[string]string{}
	inputVarsSlice := []InputVar{}
	diags := state.InputVars.ElementsAs(ctx, &inputVarsSlice, false)
	if diags.HasError() {
		return
	}
	for _, v := range inputVarsSlice {
		// store var type for later use when fetching the input variables from the API
		varTypes[v.Name.ValueString()] = v.VariableType.ValueString()
	}

	client := r.client

	addOn, err := clients.GetAddOnByID(ctx, client, loc, state.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Add-on not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Add-on", err.Error())
		return
	}

	state.ID = types.StringValue(addOn.ID)
	state.Name = types.StringValue(addOn.Name)
	state.ProjectID = types.StringValue(projectID)
	state.OrgID = types.StringValue(orgID)
	state.Summary = types.StringValue(addOn.Summary)
	state.TerraformNoCodeModuleSource = types.StringValue(addOn.TerraformNocodeModule.Source)

	state.Description = types.StringValue(addOn.Description)
	// set plan.description if it's not null or addOn.description is not empty
	if addOn.Description == "" {
		state.Description = types.StringNull()
	}
	state.ReadmeMarkdown = types.StringValue(addOn.ReadmeMarkdown.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdown.String() == "" {
		state.ReadmeMarkdown = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Labels = labels

	state.CreatedBy = types.StringValue(addOn.CreatedBy)

	// If we can process status as an int64, add it to the plan
	statusNum, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-on status", err.Error())
	} else {
		state.Status = types.Int64Value(statusNum)
	}

	// If we can process count as an int64, add it to the plan
	installedCount, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-ons count", err.Error())
	} else {
		state.Count = types.Int64Value(installedCount)
	}

	// Display the reference to the Definition in the state
	if addOn.Definition != nil {
		if addOn.Definition.ID != "" {
			state.DefinitionID = types.StringValue(addOn.Definition.ID)
		}
	}

	// Display the reference to the Application in the state
	if addOn.Application != nil {
		if addOn.Application.ID != "" {
			state.ApplicationID = types.StringValue(addOn.Application.ID)
		}
	}

	ol := readOutputs(addOn.OutputValues)
	if len(ol) > 0 {
		state.OutputValues, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: outputValue{}.attrTypes()}, ol)
	} else {
		state.OutputValues = types.ListNull(types.ObjectType{AttrTypes: outputValue{}.attrTypes()})
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputVars, err := clients.GetInputVariables(ctx, client, state.Name.ValueString(), loc)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to fetch application's input variables.")
		return
	}

	applicationInputVars, templateInputVars := splitInputs(inputVars, varTypes)
	if len(applicationInputVars) > 0 {
		aivs, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: InputVar{}.attrTypes()}, applicationInputVars)
		if diags.HasError() {
			return
		}
		state.InputVars = aivs
	} else {
		state.InputVars = types.SetNull(types.ObjectType{AttrTypes: InputVar{}.attrTypes()})
	}

	if len(templateInputVars) > 0 {
		tivs, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: InputVar{}.attrTypes()}, templateInputVars)
		if diags.HasError() {
			return
		}
		state.AddOnDefinitionInputVars = tivs
	} else {
		state.AddOnDefinitionInputVars = types.SetNull(types.ObjectType{AttrTypes: InputVar{}.attrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AddOnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *AddOnResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID

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

	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateAddOnBody{
		Name: plan.Name.ValueString(),
	}

	params := &waypoint_service.WaypointServiceUpdateAddOnParams{
		NamespaceID:     ns.ID,
		Body:            modelBody,
		ExistingAddOnID: plan.ID.ValueString(),
	}
	def, err := r.client.Waypoint.WaypointServiceUpdateAddOn(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Add-on", err.Error())
		return
	}

	var addOn *waypoint_models.HashicorpCloudWaypointAddOn
	if def.Payload != nil {
		addOn = def.Payload.AddOn
	}
	if addOn == nil {
		resp.Diagnostics.AddError("Unknown error updating Add-on", "Empty Add-on found")
		return
	}

	plan.ID = types.StringValue(addOn.ID)
	plan.Name = types.StringValue(addOn.Name)
	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOn.Summary)
	plan.TerraformNoCodeModuleSource = types.StringValue(addOn.TerraformNocodeModule.Source)

	plan.Description = types.StringValue(addOn.Description)
	// set plan.description if it's not null or addOn.description is not empty
	if addOn.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdown = types.StringValue(addOn.ReadmeMarkdown.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdown.String() == "" {
		plan.ReadmeMarkdown = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	// Display the reference to the Definition in the plan
	if addOn.Definition != nil {
		if addOn.Definition.ID != "" {
			plan.DefinitionID = types.StringValue(addOn.Definition.ID)
		}
	}

	// Display the reference to the Application in the plan
	if addOn.Application != nil {
		if addOn.Application.ID != "" {
			plan.ApplicationID = types.StringValue(addOn.Application.ID)
		}
	}

	plan.CreatedBy = types.StringValue(addOn.CreatedBy)

	// If we can process status as an int64, add it to the plan
	statusNum, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-on status", err.Error())
	} else {
		plan.Status = types.Int64Value(statusNum)
	}

	// If we can process count as an int64, add it to the plan
	installedCount, err := strconv.ParseInt(addOn.Count, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing installed Add-ons count", err.Error())
	} else {
		plan.Count = types.Int64Value(installedCount)
	}

	ol := readOutputs(addOn.OutputValues)
	if len(ol) > 0 {
		plan.OutputValues, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: outputValue{}.attrTypes()}, ol)
	} else {
		plan.OutputValues = types.ListNull(types.ObjectType{AttrTypes: outputValue{}.attrTypes()})
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated add-on resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AddOnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *AddOnResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID

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

	params := &waypoint_service.WaypointServiceDestroyAddOnParams{
		NamespaceID: ns.ID,
		AddOnID:     state.ID.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDestroyAddOn(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Add-on not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Add-on",
			err.Error(),
		)
		return
	}

}

func (r *AddOnResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readOutputs accepts a list of output values in the type returned by the Waypoint API and returns a list of output
// values in the custom outputValue type used in this provider
func readOutputs(ovs []*waypoint_models.HashicorpCloudWaypointTFOutputValue) []*outputValue {
	ol := make([]*outputValue, len(ovs))
	for i, ov := range ovs {
		ol[i] = &outputValue{
			Name:      types.StringValue(ov.Name),
			Type:      types.StringValue(ov.Type),
			Value:     types.StringValue(ov.Value),
			Sensitive: types.BoolValue(ov.Sensitive),
		}
	}
	return ol
}
