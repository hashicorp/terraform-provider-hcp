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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			/*"created_at": schema.StringAttribute{
				Description: "The timestamp when the Action Config was created in the database.",
				Computed:    true,
			},*/
			"request": schema.ListNestedAttribute{
				Description: "The kind of HTTP request this config should trigger.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"custom": schema.ListNestedAttribute{
							Description: "Custom mode allows users to define the HTTP method, the request body, etc.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"method": schema.StringAttribute{
										Description: "The HTTP method to use for the request.",
										Computed:    true,
									},
									"headers": schema.MapAttribute{
										Description: "Key value headers to send with the request.",
										Computed:    true,
									},
									"url": schema.StringAttribute{
										Description: "The full URL this request should make when invoked.",
										Computed:    true,
									},
									"body": schema.StringAttribute{
										Description: "The body to be submitted with the request.",
										Computed:    true,
									},
								},
							},
						},
						/*"github": schema.ListNestedAttribute{
							Description: "GitHub mode is configured to do various operations on GitHub Repositories.",
							Optional:    true,
						},
						"agent": schema.ListNestedAttribute{
							Optional: true,
						},*/
					},
				},
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

	actionCfg, err = clients.GetActionConfig(ctx, client, loc, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// TODO: Add remaining data types
	data.ID = types.StringValue(actionCfg.ID)
	data.ActionURL = types.StringValue(actionCfg.ActionURL)

	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)

	data.Request = actionConfigRequest{}
	headerMap := make(map[types.String][]types.String)

	var diags diag.Diagnostics

	// In the future, expand this to accommodate other types of requests
	if actionCfg.Request.Custom != nil {
		data.Request.custom = customRequest{}
		if actionCfg.Request.Custom.Method != nil {
			methodString, err := ConvertMethodToStringType(*actionCfg.Request.Custom.Method)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unexpected HTTP Method",
					"Expected GET, POST, PUT, DELETE, or PATCH. Please report this issue to the provider developers.",
				)
			} else {
				data.Request.custom.Method = methodString
			}
		}
		if actionCfg.Request.Custom.Headers != nil {
			for _, header := range actionCfg.Request.Custom.Headers {
				headerMap[types.StringValue(header.Key)] = append(headerMap[types.StringValue(header.Key)], types.StringValue(header.Value))
			}
			data.Request.custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if actionCfg.Request.Custom.URL != "" {
			data.Request.custom.URL = types.StringValue(actionCfg.Request.Custom.URL)
		}
		if actionCfg.Request.Custom.Body != "" {
			data.Request.custom.Body = types.StringValue(actionCfg.Request.Custom.Body)
		}
	}
}
