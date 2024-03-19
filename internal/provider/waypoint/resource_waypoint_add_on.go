// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"encoding/base64"
	"fmt"

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

// TODO: Get rid of most of these because they are not used in the protos?
// AddOnResourceModel describes the resource data model.
type AddOnResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Summary        types.String `tfsdk:"summary"`
	Labels         types.List   `tfsdk:"labels"`
	Description    types.String `tfsdk:"description"`
	ReadmeMarkdown types.String `tfsdk:"readme_markdown"`
	CreatedBy      types.String `tfsdk:"created_by"`
	Count          types.Int64  `tfsdk:"count"`
	Status         types.Number `tfsdk:"status"`
	OutputValues   types.List   `tfsdk:"output_values"`

	Application           *applicationRef     `tfsdk:"application"`
	Definition            *addOnDefinitionRef `tfsdk:"definition"`
	TerraformNoCodeModule *tfcNoCodeModule    `tfsdk:"terraform_no_code_module"`
}

func (r *AddOnResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_add_on"
}

// TODO: Get rid of most of these because they are not used in the protos?
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
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A longer description of the Add-on.",
				Required:    true,
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				Description: "List of labels attached to this Add-on.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"readme_markdown": schema.StringAttribute{
				Description: "The markdown for the Add-on README.",
				Optional:    true,
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
						Description: "Tetraform Cloud Project ID",
					},
				},
			},
			"terraform_no_code_module": &schema.SingleNestedAttribute{
				Required:    true,
				Description: "Terraform Cloud no-code Module details.",
				Attributes: map[string]schema.Attribute{
					"source": &schema.StringAttribute{
						Required:    true,
						Description: "Terraform Cloud no-code Module Source",
					},
					"version": &schema.StringAttribute{
						Required:    true,
						Description: "Terraform Cloud no-code Module Version",
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
	// so if we have both, we'll use ID
	applicationId := plan.Application.ID.ValueString()
	applicationName := plan.Application.Name.ValueString()
	applicationRefModel := &waypointmodels.HashicorpCloudWaypointRefApplication{}
	if applicationId != "" {
		applicationRefModel.ID = applicationId
	} else if applicationName != "" {
		applicationRefModel.Name = applicationName
	} else {
		resp.Diagnostics.AddError(
			"error reading application ref",
			"The application reference was missing",
		)
		return
	}

	// Similarly, a definition ref can only have one of ID or Name set,
	// so if we have both, we'll use ID
	definitionId := plan.Definition.ID.ValueString()
	definitionName := plan.Definition.Name.ValueString()
	definitionRefModel := &waypointmodels.HashicorpCloudWaypointRefAddOnDefinition{}
	if definitionId != "" {
		definitionRefModel.ID = definitionId
	} else if definitionName != "" {
		definitionRefModel.Name = definitionName
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
	def, err := r.client.Waypoint.WaypointServiceCreateAddOn(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating add-on", err.Error())
		return
	}

	var addOn *waypointmodels.HashicorpCloudWaypointAddOn
	if def.Payload != nil {
		addOn = def.Payload.AddOn
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
	plan.ReadmeMarkdownTemplate = types.StringValue(addOn.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	if addOn.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOn.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOn.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace = tfcWorkspace
	}

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule = tfcNoCode
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
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client

	addOn, err := clients.GetAddOnnByID(ctx, client, loc, state.ID.ValueString())
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
	state.OrgID = types.StringValue(client.Config.OrganizationID)
	state.ProjectID = types.StringValue(client.Config.ProjectID)
	state.Summary = types.StringValue(addOn.Summary)

	state.Description = types.StringValue(addOn.Description)
	// set plan.description if it's not null or addOn.description is not empty
	if addOn.Description == "" {
		state.Description = types.StringNull()
	}
	state.ReadmeMarkdownTemplate = types.StringValue(addOn.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdownTemplate.String() == "" {
		state.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Labels = labels

	if addOn.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOn.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOn.TerraformCloudWorkspaceDetails.ProjectID),
		}
		state.TerraformCloudWorkspace = tfcWorkspace
	}

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		state.TerraformNoCodeModule = tfcNoCode
	}

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

	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"error decoding base64 readme markdown template",
			err.Error(),
		)
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
	// TODO: add support for Tags
	modelBody := &waypointmodels.HashicorpCloudWaypointWaypointServiceUpdateAddOnBody{
		Name:                   plan.Name.ValueString(),
		Summary:                plan.Summary.ValueString(),
		Description:            plan.Description.ValueString(),
		ReadmeMarkdownTemplate: readmeBytes,
		Labels:                 stringLabels,
		TerraformNocodeModule: &waypointmodels.HashicorpCloudWaypointTerraformNocodeModule{
			// verify these exist in the file
			Source:  plan.TerraformNoCodeModule.Source.ValueString(),
			Version: plan.TerraformNoCodeModule.Version.ValueString(),
		},
		TerraformCloudWorkspaceDetails: &waypointmodels.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
			ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
		},
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
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(addOn.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOn.Summary)

	plan.Description = types.StringValue(addOn.Description)
	// set plan.description if it's not null or addOn.description is not empty
	if addOn.Description == "" {
		plan.Description = types.StringNull()
	}
	plan.ReadmeMarkdownTemplate = types.StringValue(addOn.ReadmeMarkdownTemplate.String())
	// set plan.readme if it's not null or addOn.readme is not empty
	if addOn.ReadmeMarkdownTemplate.String() == "" {
		plan.ReadmeMarkdownTemplate = types.StringNull()
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Labels = labels

	if addOn.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOn.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOn.TerraformCloudWorkspaceDetails.ProjectID),
		}
		plan.TerraformCloudWorkspace = tfcWorkspace
	}

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule = tfcNoCode
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

	params := &waypoint_service.WaypointServiceDeleteAddOnParams{
		NamespaceID: ns.ID,
		AddOnID:     state.ID.ValueString(),
	}

	_, err = r.client.Waypoint.WaypointServiceDeleteAddOn(params, nil)
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
