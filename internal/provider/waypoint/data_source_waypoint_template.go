// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ datasource.DataSource = &DataSourceTemplate{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceTemplate{}

func (d DataSourceTemplate) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceTemplate struct {
	client *clients.Client
}

type DataSourceTemplateModel struct {
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
	VariableOptions             []*tfcVariableOption `tfsdk:"variable_options"`
}

func NewTemplateDataSource() datasource.DataSource {
	return &DataSourceTemplate{}
}

func (d *DataSourceTemplate) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_template"
}

func (d *DataSourceTemplate) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Template data source retrieves information on a given Template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Template.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Template.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Template is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Template is located.",
				Optional:    true,
				Computed:    true,
			},
			"summary": schema.StringAttribute{
				Description: "A brief description of the template, up to 110 characters",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the template, along with when and why it should be used, up to 500 characters",
			},
			"readme_markdown_template": schema.StringAttribute{
				Computed:    true,
				Description: "Instructions for using the template (markdown format supported)",
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Description: "List of labels attached to this Template.",
				ElementType: types.StringType,
			},
			"terraform_cloud_workspace_details": &schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Terraform Cloud Workspace details",
				Attributes: map[string]schema.Attribute{
					"name": &schema.StringAttribute{
						Computed:    true,
						Description: "Name of the Terraform Cloud Workspace",
					},
					"terraform_project_id": &schema.StringAttribute{
						Computed:    true,
						Description: "Terraform Cloud Project ID",
					},
				},
			},
			"terraform_no_code_module_source": &schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Terraform No Code Module source",
			},
			"variable_options": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of variable options for the template",
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
							Description: "Whether the variable is editable by the user creating an application",
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceTemplate) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceTemplate) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceTemplateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	client := d.client
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	projectID := client.Config.ProjectID
	if !data.ProjectID.IsNull() {
		projectID = data.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	var appTemplate *waypoint_models.HashicorpCloudWaypointApplicationTemplate
	var err error

	if data.ID.IsNull() {
		appTemplate, err = clients.GetApplicationTemplateByName(ctx, client, loc, data.Name.ValueString())
	} else if data.Name.IsNull() {
		appTemplate, err = clients.GetApplicationTemplateByID(ctx, client, loc, data.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	data.ID = types.StringValue(appTemplate.ID)
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

	data.VariableOptions, err = readVarOpts(ctx, appTemplate.VariableOptions, &resp.Diagnostics)
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
