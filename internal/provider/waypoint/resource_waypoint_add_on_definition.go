// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-openapi/strfmt"

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

	TerraformCloudWorkspace *tfcWorkspace    `tfsdk:"terraform_cloud_workspace_details"`
	TerraformNoCodeModule   *tfcNoCodeModule `tfsdk:"terraform_no_code_module"`

	// questionable
	// Namespace types.String `tfsdk:"namespace_id"`
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
				Description: "The ID of the HCP project where the Waypoint Add-on Definition is located.",
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
				Description: "The markdown template for the Add-on Definition README.",
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
				Description: "Terraform Cloud No Code Module details",
				Attributes: map[string]schema.Attribute{
					"source": &schema.StringAttribute{
						Required:    true,
						Description: "No Code Module Source",
					},
					"version": &schema.StringAttribute{
						Required:    true,
						Description: "No Code Module Version",
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

	// TODO: (Henry) Follow up, is this the best way to type convert here?
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
	readmeBytes, err := base64.StdEncoding.DecodeString(plan.ReadmeMarkdownTemplate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"error decoding base64 readme markdown template",
			err.Error(),
		)
	}
	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceCreateAddOnDefinitionBody{
		Name:                   plan.Name.ValueString(),
		Summary:                plan.Summary.ValueString(),
		Description:            plan.Description.ValueString(),
		ReadmeMarkdownTemplate: readmeBytes,
		Labels:                 stringLabels,
		TerraformNocodeModule: &waypoint_models.HashicorpCloudWaypointTerraformNocodeModule{
			// verify these exist in the file
			Source:  plan.TerraformNoCodeModule.Source.ValueString(),
			Version: plan.TerraformNoCodeModule.Version.ValueString(),
		},
		TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      plan.TerraformCloudWorkspace.Name.ValueString(),
			ProjectID: plan.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
		},
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

	var addOnDefinition *waypoint_models.HashicorpCloudWaypointAddOnDefinition
	if def.Payload != nil {
		addOnDefinition = def.Payload.AddOnDefinition
	}
	if addOnDefinition == nil {
		resp.Diagnostics.AddError("unknown error creating add-on definition", "empty add-on definition found")
		return
	}

	plan.ID = types.StringValue(addOnDefinition.ID)
	plan.ProjectID = types.StringValue(projectID)
	plan.Name = types.StringValue(addOnDefinition.Name)
	plan.OrgID = types.StringValue(orgID)
	plan.Summary = types.StringValue(addOnDefinition.Summary)
	plan.Description = types.StringValue(addOnDefinition.Description)
	plan.ReadmeMarkdownTemplate = types.StringValue(addOnDefinition.ReadmeMarkdownTemplate.String())
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
		plan.TerraformCloudWorkspace = tfcWorkspace
	}

	if addOnDefinition.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOnDefinition.TerraformNocodeModule.Source),
			Version: types.StringValue(addOnDefinition.TerraformNocodeModule.Version),
		}
		plan.TerraformNoCodeModule = tfcNoCode
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created add-on definition resource")

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AddOnDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AddOnDefinitionResourceModel

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

	definition, err := clients.GetAddOnDefinitionByID(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "TFC Config not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading TFC Config", err.Error())
		return
	}

	data.ID = types.StringValue(definition.ID)
	data.Name = types.StringValue(definition.Name)
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	data.Summary = types.StringValue(definition.Summary)
	data.Description = types.StringValue(definition.Description)
	data.ReadmeMarkdownTemplate = types.StringValue(definition.ReadmeMarkdownTemplate.String())

	labels, diags := types.ListValueFrom(ctx, types.StringType, definition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Labels = labels

	if definition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(definition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(definition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		data.TerraformCloudWorkspace = tfcWorkspace
	}

	if definition.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(definition.TerraformNocodeModule.Source),
			Version: types.StringValue(definition.TerraformNocodeModule.Version),
		}
		data.TerraformNoCodeModule = tfcNoCode
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AddOnDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AddOnDefinitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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
	ns, err := getNamespaceByLocation(ctx, client, loc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting namespace by location",
			err.Error(),
		)
		return
	}

	// TODO: (Henry) add support for Labels and Tags
	modelBody := &waypoint_models.HashicorpCloudWaypointWaypointServiceUpdateAddOnDefinitionBody{
		Name:                   data.Name.ValueString(),
		Summary:                data.Summary.ValueString(),
		Description:            data.Description.ValueString(),
		ReadmeMarkdownTemplate: strfmt.Base64(data.ReadmeMarkdownTemplate.ValueString()),
		TerraformNocodeModule: &waypoint_models.HashicorpCloudWaypointTerraformNocodeModule{
			// verify these exist in the file
			Source:  data.TerraformNoCodeModule.Source.ValueString(),
			Version: data.TerraformNoCodeModule.Version.ValueString(),
		},
		TerraformCloudWorkspaceDetails: &waypoint_models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      data.TerraformCloudWorkspace.Name.ValueString(),
			ProjectID: data.TerraformCloudWorkspace.TerraformProjectID.ValueString(),
		},
	}

	params := &waypoint_service.WaypointServiceUpdateAddOnDefinitionParams{
		NamespaceID:               ns.ID,
		Body:                      modelBody,
		ExistingAddOnDefinitionID: data.ID.ValueString(),
	}
	def, err := r.client.Waypoint.WaypointServiceUpdateAddOnDefinition(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	var addOnDefinition *waypoint_models.HashicorpCloudWaypointAddOnDefinition
	if def.Payload != nil {
		addOnDefinition = def.Payload.AddOnDefinition
	}
	if addOnDefinition == nil {
		resp.Diagnostics.AddError("unknown error updating add-on definition", "empty add-on definition found")
		return
	}

	data.ID = types.StringValue(addOnDefinition.ID)
	data.ProjectID = types.StringValue(projectID)
	data.Name = types.StringValue(addOnDefinition.Name)
	data.OrgID = types.StringValue(orgID)
	data.Summary = types.StringValue(addOnDefinition.Summary)
	data.Description = types.StringValue(addOnDefinition.Description)
	data.ReadmeMarkdownTemplate = types.StringValue(addOnDefinition.ReadmeMarkdownTemplate.String())

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOnDefinition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Labels = labels

	if addOnDefinition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(addOnDefinition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		data.TerraformCloudWorkspace = tfcWorkspace
	}

	if addOnDefinition.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOnDefinition.TerraformNocodeModule.Source),
			Version: types.StringValue(addOnDefinition.TerraformNocodeModule.Version),
		}
		data.TerraformNoCodeModule = tfcNoCode
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "updated add-on definition resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AddOnDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AddOnDefinitionResourceModel

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

	params := &waypoint_service.WaypointServiceDeleteAddOnDefinitionParams{
		NamespaceID:       ns.ID,
		AddOnDefinitionID: data.ID.ValueString(),
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
