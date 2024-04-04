// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"strconv"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	waypointmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
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
	Summary        types.String `tfsdk:"summary"`
	Labels         types.List   `tfsdk:"labels"`
	Description    types.String `tfsdk:"description"`
	ReadmeMarkdown types.String `tfsdk:"readme_markdown"`
	CreatedBy      types.String `tfsdk:"created_by"`
	Count          types.Int64  `tfsdk:"install_count"`
	Status         types.Int64  `tfsdk:"status"`
	//OutputValues   types.List   `tfsdk:"output_values"`

	Application           *addOnApplicationRef `tfsdk:"application"`
	Definition            *addOnDefinitionRef  `tfsdk:"definition"`
	TerraformNoCodeModule types.Object         `tfsdk:"terraform_no_code_module"`
}

func (r tfcNoCodeModule) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source":  types.StringType,
		"version": types.StringType,
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
			"terraform_no_code_module": &schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Terraform Cloud no-code Module details.",
				Attributes: map[string]schema.Attribute{
					"source": &schema.StringAttribute{
						Computed:    true,
						Description: "Terraform Cloud no-code Module Source",
					},
					"version": &schema.StringAttribute{
						Computed:    true,
						Description: "Terraform Cloud no-code Module Version",
					},
				},
			},
			"definition": schema.SingleNestedAttribute{
				Required: true,
				Description: "The Add-on Definition from which this Add-on was created. At least " +
					"one of ID or Name must be set.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:    true,
						Description: "The name of the Add-on Definition.",
					},
					"id": schema.StringAttribute{
						Optional:    true,
						Description: "The ID of the Add-on Definition.",
					},
				},
			},
			"application": schema.SingleNestedAttribute{
				Required: true,
				Description: "The Application for which this Add-on was created. At least one of " +
					"ID or Name must be set.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:    true,
						Description: "The name of the Application.",
					},
					"id": schema.StringAttribute{
						Optional:    true,
						Description: "The ID of the Application.",
					},
				},
			},
			"status": schema.Int64Attribute{
				Computed:    true,
				Description: "The status of the Terraform run for the Add-on.",
			},
			/*"output_values": schema.ListNestedAttribute{
				Computed: true,
				Description: "The output values of the Terraform run for the Add-on, sensitive values have type " +
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
			},*/
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

	// An application ref can only have one of ID or Name set,
	// so if we have both, we'll use Name
	applicationId := plan.Application.ID.ValueString()
	applicationName := plan.Application.Name.ValueString()
	applicationRefModel := &waypointmodels.HashicorpCloudWaypointRefApplication{}
	if applicationName != "" {
		applicationRefModel.Name = applicationName
	} else if applicationId != "" {
		applicationRefModel.ID = applicationId
	} else {
		resp.Diagnostics.AddError(
			"error reading application ref",
			"The application reference was missing",
		)
		return
	}

	// Similarly, a definition ref can only have one of ID or Name set,
	// so if we have both, we'll use Name
	definitionId := plan.Definition.ID.ValueString()
	definitionName := plan.Definition.Name.ValueString()
	definitionRefModel := &waypointmodels.HashicorpCloudWaypointRefAddOnDefinition{}
	if definitionName != "" {
		definitionRefModel.Name = definitionName
	} else if definitionId != "" {
		definitionRefModel.ID = definitionId
	} else {
		resp.Diagnostics.AddError(
			"error reading definition ref",
			"The definition reference was missing",
		)
		return
	}

	modelBody := &waypointmodels.HashicorpCloudWaypointWaypointServiceCreateAddOnBody{
		Name:        plan.Name.ValueString(),
		Application: applicationRefModel,
		Definition:  definitionRefModel,
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

	var addOn *waypointmodels.HashicorpCloudWaypointAddOn
	if responseAddOn.Payload != nil {
		addOn = responseAddOn.Payload.AddOn
	}
	if addOn == nil {
		resp.Diagnostics.AddError("unknown error creating add-on", "empty add-on returned")
		return
	}

	plan.ID = types.StringValue(addOn.ID)
	plan.Name = types.StringValue(addOn.Name)
	plan.Summary = types.StringValue(addOn.Summary)

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

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{}
		if addOn.TerraformNocodeModule.Source != "" {
			tfcNoCode.Source = types.StringValue(addOn.TerraformNocodeModule.Source)
		}
		if addOn.TerraformNocodeModule.Version != "" {
			tfcNoCode.Version = types.StringValue(addOn.TerraformNocodeModule.Version)
		}
		plan.TerraformNoCodeModule, diags = types.ObjectValueFrom(ctx, tfcNoCode.attrTypes(), tfcNoCode)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Display the reference to the Definition in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Definition != nil {
		definitionRef := &addOnDefinitionRef{}
		if addOn.Definition.Name != "" {
			definitionRef.Name = types.StringValue(addOn.Definition.Name)
		} else if addOn.Definition.ID != "" {
			definitionRef.ID = types.StringValue(addOn.Definition.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading definition ref",
				"the definition reference was missing, no ID or Name found")
			return
		}
		plan.Definition = definitionRef
	}

	// Display the reference to the Application in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Application != nil {
		applicationRef := &addOnApplicationRef{}
		if addOn.Application.Name != "" {
			applicationRef.Name = types.StringValue(addOn.Application.Name)
		} else if addOn.Application.ID != "" {
			applicationRef.ID = types.StringValue(addOn.Application.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading application ref",
				"the application reference was missing, no ID or Name found")
			return
		}
		plan.Application = applicationRef
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

	// TODO: Add support for output values

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

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
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
	state.Summary = types.StringValue(addOn.Summary)

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

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		state.TerraformNoCodeModule, diags = types.ObjectValueFrom(ctx, tfcNoCode.attrTypes(), tfcNoCode)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.CreatedBy = types.StringValue(addOn.CreatedBy)

	// TODO: Error out here on failure to convert?

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

	// Display the reference to the Definition in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Definition != nil {
		definitionRef := &addOnDefinitionRef{}
		if addOn.Definition.Name != "" {
			definitionRef.Name = types.StringValue(addOn.Definition.Name)
		} else if addOn.Definition.ID != "" {
			definitionRef.ID = types.StringValue(addOn.Definition.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading definition ref",
				"the definition reference was missing, no ID or Name found")
			return
		}
		state.Definition = definitionRef
	}

	// Display the reference to the Application in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Application != nil {
		applicationRef := &addOnApplicationRef{}
		if addOn.Application.Name != "" {
			applicationRef.Name = types.StringValue(addOn.Application.Name)
		} else if addOn.Application.ID != "" {
			applicationRef.ID = types.StringValue(addOn.Application.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading application ref",
				"the application reference was missing, no ID or Name found")
			return
		}
		state.Application = applicationRef
	}

	// TODO: Add support for output values

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TODO: Add support for new fields
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

	modelBody := &waypointmodels.HashicorpCloudWaypointWaypointServiceUpdateAddOnBody{
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

	var addOn *waypointmodels.HashicorpCloudWaypointAddOn
	if def.Payload != nil {
		addOn = def.Payload.AddOn
	}
	if addOn == nil {
		resp.Diagnostics.AddError("Unknown error updating Add-on", "Empty Add-on found")
		return
	}

	plan.ID = types.StringValue(addOn.ID)
	plan.Name = types.StringValue(addOn.Name)
	plan.Summary = types.StringValue(addOn.Summary)

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

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule, diags = types.ObjectValueFrom(ctx, tfcNoCode.attrTypes(), tfcNoCode)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Display the reference to the Definition in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Definition != nil {
		definitionRef := &addOnDefinitionRef{}
		if addOn.Definition.Name != "" {
			definitionRef.Name = types.StringValue(addOn.Definition.Name)
		} else if addOn.Definition.ID != "" {
			definitionRef.ID = types.StringValue(addOn.Definition.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading definition ref",
				"the definition reference was missing, no ID or Name found")
			return
		}
		plan.Definition = definitionRef
	}

	// Display the reference to the Application in the plan,
	// prioritizing displaying Name because it is more user-friendly
	if addOn.Application != nil {
		applicationRef := &addOnApplicationRef{}
		if addOn.Application.Name != "" {
			applicationRef.Name = types.StringValue(addOn.Application.Name)
		} else if addOn.Application.ID != "" {
			applicationRef.ID = types.StringValue(addOn.Application.ID)
		} else {
			// One of the above cases has to be true, this should not happen
			resp.Diagnostics.AddError(
				"error reading application ref",
				"the application reference was missing, no ID or Name found")
			return
		}
		plan.Application = applicationRef
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

	// TODO: Add support for output values

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
