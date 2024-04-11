// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ datasource.DataSource = &DataSourceActionConfig{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceActionConfig{}

func (d DataSourceActionConfig) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceActionConfig struct {
	client *clients.Client
}

func NewActionConfigDataSource() datasource.DataSource {
	return &DataSourceActionConfig{}
}

func (d *DataSourceActionConfig) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_action_config"
}

// TODO: Add remaining data types
func (d *DataSourceActionConfig) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Action Config data source retrieves information on a given Action Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Action Config.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Action Config.",
				Computed:    true,
				Optional:    true,
			},
			"namespace_id": schema.StringAttribute{
				Description: "The ID of the namespace where the Waypoint Action Config is located.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Action Config is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Action Config is located.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the Action Config.",
				Computed:    true,
			},
			"action_url": schema.StringAttribute{
				Description: "The URL to trigger an action on. Only used in Custom mode",
				Computed:    true,
			},
		},
	}
}

func (d *DataSourceActionConfig) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceActionConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionConfigResourceModel

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

	var actionCfg *waypoint_models.HashicorpCloudWaypointActionConfig
	var err error

	if data.ID.IsNull() {
		actionCfg, err = clients.GetActionConfigByName(ctx, client, loc, data.Name.ValueString())
	} else if data.Name.IsNull() {
		actionCfg, err = clients.GetActionConfigByID(ctx, client, loc, data.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// TODO: Add remaining data types
	data.ID = types.StringValue(actionCfg.ID)
	data.ActionURL = types.StringValue(actionCfg.ActionURL)

	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
}
