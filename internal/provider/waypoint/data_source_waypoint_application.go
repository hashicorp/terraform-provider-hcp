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

var _ datasource.DataSource = &DataSourceApplication{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceApplication{}

func (d DataSourceApplication) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceApplication struct {
	client *clients.Client
}

func NewApplicationDataSource() datasource.DataSource {
	return &DataSourceApplication{}
}

func (d *DataSourceApplication) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_application"
}

func (d *DataSourceApplication) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Application data source retrieves information on a given Application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Application.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Application.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Application is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Application is located.",
				Optional:    true,
				Computed:    true,
			},
			"readme_markdown": schema.StringAttribute{
				Computed:    true,
				Description: "Instructions for using the Application (markdown format supported).",
			},
			"application_template_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the Application Template this Application is based on.",
			},
			"application_template_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Application Template this Application is based on.",
			},
			"namespace_id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal Namespace ID.",
			},
			"input_vars": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Input variables for the Application.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": &schema.StringAttribute{
							Computed:    true,
							Description: "Variable name",
						},
						"value": &schema.StringAttribute{
							Computed:    true,
							Description: "Variable value",
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceApplication) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceApplication) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationResourceModel
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

	var application *waypoint_models.HashicorpCloudWaypointApplication
	var err error

	if data.ID.IsNull() {
		application, err = clients.GetApplicationByName(ctx, client, loc, data.Name.ValueString())
	} else if data.Name.IsNull() {
		application, err = clients.GetApplicationByID(ctx, client, loc, data.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	data.ID = types.StringValue(application.ID)
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	// set data.readme if it's not null or application.readme is not
	// empty
	data.ReadmeMarkdown = types.StringValue(application.ReadmeMarkdown.String())
	if application.ReadmeMarkdown.String() == "" {
		data.ReadmeMarkdown = types.StringNull()
	}

	// A second API call is made to get the input vars set on the application
	inputVars, err := clients.GetInputVariables(ctx, client, data.Name.ValueString(), loc)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to fetch application's input variables.")
		return
	}

	for _, iv := range inputVars {
		data.InputVars = append(data.InputVars, &InputVar{
			Name:         types.StringValue(iv.Name),
			Value:        types.StringValue(iv.Value),
			VariableType: types.StringValue(iv.VariableType),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
