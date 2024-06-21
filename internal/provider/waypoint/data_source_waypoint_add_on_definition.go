// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	waypointModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ datasource.DataSource = &DataSourceAddOnDefinition{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceAddOnDefinition{}

func (d DataSourceAddOnDefinition) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceAddOnDefinition struct {
	client *clients.Client
}

type DataSourceAddOnDefinitionModel struct {
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

func NewAddOnDefinitionDataSource() datasource.DataSource {
	return &DataSourceAddOnDefinition{}
}

func (d *DataSourceAddOnDefinition) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_add_on_definition"
}

func (d *DataSourceAddOnDefinition) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Add-on Definition data source retrieves information on a given Add-on Definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Add-on Definition.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Add-on Definition.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Add-on Definition is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Add-on Definition is located.",
				Optional:    true,
				Computed:    true,
			},
			"summary": schema.StringAttribute{
				Description: "A short summary of the Add-on Definition.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A longer description of the Add-on Definition.",
				Computed:    true,
			},
			"readme_markdown_template": schema.StringAttribute{
				Computed:    true,
				Description: "Instructions for using the Add-on Definition (markdown format supported).",
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Description: "List of labels attached to this Add-on Definition.",
				ElementType: types.StringType,
			},
			"terraform_cloud_workspace_details": &schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Terraform Cloud Workspace details.",
				Attributes: map[string]schema.Attribute{
					"name": &schema.StringAttribute{
						Computed:    true,
						Description: "Name of the Terraform Cloud Workspace.",
					},
					"terraform_project_id": &schema.StringAttribute{
						Computed:    true,
						Description: "Terraform Cloud Project ID.",
					},
				},
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
			"variable_options": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of variable options for the Add-on Definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": &schema.StringAttribute{
							Computed:    true,
							Description: "Variable name",
						},
						"variable_type": &schema.StringAttribute{
							Computed:    true,
							Description: "Variable type",
						},
						"options": &schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "List of options",
						},
						"user_editable": &schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the variable is editable by the user creating an add-on",
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceAddOnDefinition) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DataSourceAddOnDefinition) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceAddOnDefinitionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	client := d.client
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	projectID := client.Config.ProjectID
	if !state.ProjectID.IsUnknown() && !state.ProjectID.IsNull() {
		projectID = state.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	var definition *waypointModels.HashicorpCloudWaypointAddOnDefinition
	var err error

	if state.ID.IsNull() {
		definition, err = clients.GetAddOnDefinitionByName(ctx, client, loc, state.Name.ValueString())
	} else if state.Name.IsNull() {
		definition, err = clients.GetAddOnDefinitionByID(ctx, client, loc, state.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	state.ID = types.StringValue(definition.ID)
	state.OrgID = types.StringValue(client.Config.OrganizationID)
	state.ProjectID = types.StringValue(client.Config.ProjectID)
	state.Summary = types.StringValue(definition.Summary)
	state.Description = types.StringValue(definition.Description)

	if definition.TerraformCloudWorkspaceDetails != nil {
		tfcWorkspace := &tfcWorkspace{
			Name:               types.StringValue(definition.TerraformCloudWorkspaceDetails.Name),
			TerraformProjectID: types.StringValue(definition.TerraformCloudWorkspaceDetails.ProjectID),
		}
		state.TerraformCloudWorkspace = tfcWorkspace
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, definition.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(labels.Elements()) == 0 {
		labels = types.ListNull(types.StringType)
	}
	state.Labels = labels

	// set plan.description if it's not null or addOnDefinition.description is not empty
	state.Description = types.StringValue(definition.Description)
	if definition.Description == "" {
		state.Description = types.StringNull()
	}
	state.ReadmeMarkdownTemplate = types.StringValue(definition.ReadmeMarkdownTemplate.String())
	// set state.readme if it's not null or addOnDefinition.readme is not empty
	if definition.ReadmeMarkdownTemplate.String() == "" {
		state.ReadmeMarkdownTemplate = types.StringNull()
	}

	state.TerraformVariableOptions, err = readVarOpts(ctx, definition.VariableOptions, &resp.Diagnostics)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
