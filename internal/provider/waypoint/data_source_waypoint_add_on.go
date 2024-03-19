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
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ datasource.DataSource = &DataSourceAddOn{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceAddOn{}

func (d DataSourceAddOn) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceAddOn struct {
	client *clients.Client
}

//TODO: Make sure this list is correct (is this the right type for output values?)

type DataSourceAddOnModel struct {
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

type addOnDefinitionRef struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type applicationRef struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewAddOnDataSource() datasource.DataSource {
	return &DataSourceAddOn{}
}

func (d *DataSourceAddOn) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_add_on_definition"
}

//TODO: Make sure this schema is correct (do we want to include count or output values?)

func (d *DataSourceAddOn) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Add-on data source retrieves information on a given Add-on.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Add-on.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Add-on.",
				Computed:    true,
				Optional:    true,
			},
			"summary": schema.StringAttribute{
				Description: "A short summary of the Add-on.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A longer description of the Add-on.",
				Computed:    true,
			},
			"readme_markdown": schema.StringAttribute{
				Computed:    true,
				Description: "Instructions for using the Add-on (markdown format supported).",
			},
			"labels": schema.ListAttribute{
				Computed:    true,
				Description: "List of labels attached to this Add-on.",
				ElementType: types.StringType,
			},
			"created_by": schema.StringAttribute{
				Description: "The user who created the Add-on.",
				Computed:    true,
			},
			"count": schema.Int64Attribute{
				Description: "The number of installed Add-ons for the same Application that share the same " +
					"Add-on Definition.",
				Computed: true,
			},
			"definition": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The Add-on Definition from which this Add-on was created.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the Add-on Definition.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the Add-on Definition.",
					},
				},
			},
			"application": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The Application for which this Add-on was created.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the Application.",
					},
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the Application.",
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
			"status": schema.NumberAttribute{
				Computed:    true,
				Description: "The status of the Terraform run for the Add-on.",
			},
			"output_values": schema.ListNestedAttribute{
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
			},
		},
	}
}

func (d *DataSourceAddOn) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// TODO: Convert count from uint64 to int64, add support for application, definition, created at, status, created by, (output values?)
func (d *DataSourceAddOn) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceAddOnModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	client := d.client
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	var addOn *waypoint_models.HashicorpCloudWaypointAddOn
	var err error

	if state.ID.IsNull() {
		addOn, err = clients.GetAddOnByName(ctx, client, loc, state.Name.ValueString())
	} else if state.Name.IsNull() {
		addOn, err = clients.GetAddOnByID(ctx, client, loc, state.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	state.ID = types.StringValue(addOn.ID)
	state.Summary = types.StringValue(addOn.Summary)

	if addOn.TerraformNocodeModule != nil {
		tfcNoCode := &tfcNoCodeModule{
			Source:  types.StringValue(addOn.TerraformNocodeModule.Source),
			Version: types.StringValue(addOn.TerraformNocodeModule.Version),
		}
		state.TerraformNoCodeModule = tfcNoCode
	}

	labels, diags := types.ListValueFrom(ctx, types.StringType, addOn.Labels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(labels.Elements()) == 0 {
		labels = types.ListNull(types.StringType)
	}
	state.Labels = labels

	// set plan.description if it's not null or addOn.description is not empty
	state.Description = types.StringValue(addOn.Description)
	if addOn.Description == "" {
		state.Description = types.StringNull()
	}
	state.ReadmeMarkdown = types.StringValue(addOn.ReadmeMarkdown.String())
	// set state.readme if it's not null or addOnDefinition.readme is not empty
	if addOn.ReadmeMarkdown.String() == "" {
		state.ReadmeMarkdown = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
