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

var _ datasource.DataSource = &DataSourceAction{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceAction{}

func (d DataSourceAction) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("id"),
		),
	}
}

type DataSourceAction struct {
	client *clients.Client
}

func NewActionDataSource() datasource.DataSource {
	return &DataSourceAction{}
}

func (d *DataSourceAction) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_action"
}

func (d *DataSourceAction) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Waypoint Action data source retrieves information on a given Action.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the Action.",
			},
			"name": schema.StringAttribute{
				Description: "The name of the Action.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Waypoint Action is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Waypoint Action is located.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the Action.",
				Computed:    true,
			},
			"request": schema.SingleNestedAttribute{
				Description: "The kind of HTTP request this should trigger.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"custom": schema.SingleNestedAttribute{
						Description: "Custom mode allows users to define the HTTP method, the request body, etc.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"method": schema.StringAttribute{
								Description: "The HTTP method to use for the request.",
								Computed:    true,
							},
							"headers": schema.MapAttribute{
								Description: "Key value headers to send with the request.",
								Computed:    true,
								ElementType: types.StringType,
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
			},
		},
	}
}

func (d *DataSourceAction) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceAction) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionResourceModel

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

	var actionModel *waypoint_models.HashicorpCloudWaypointActionConfig
	var err error

	actionModel, err = clients.GetAction(ctx, client, loc, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Failed to find action by ID or name")
		return
	}

	if actionModel.ID != "" {
		data.ID = types.StringValue(actionModel.ID)
	}
	if actionModel.Name != "" {
		data.Name = types.StringValue(actionModel.Name)
	}
	if actionModel.Description != "" {
		data.Description = types.StringValue(actionModel.Description)
	}

	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)

	data.Request = &actionRequest{}
	headerMap := make(map[string]string)

	var diags diag.Diagnostics

	// In the future, expand this to accommodate other types of requests

	data.Request.Custom = &customRequest{}
	if actionModel.Request.Custom.Method != nil {
		methodString, err := convertMethodToStringType(*actionModel.Request.Custom.Method)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unexpected HTTP Method",
				"Expected GET, POST, PUT, DELETE, or PATCH. Please report this issue to the provider developers.",
			)
			return
		} else {
			data.Request.Custom.Method = methodString
		}
	}
	if actionModel.Request.Custom.Headers != nil {
		for _, header := range actionModel.Request.Custom.Headers {
			headerMap[header.Key] = header.Value
		}
		if len(headerMap) > 0 {
			data.Request.Custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			data.Request.Custom.Headers = types.MapNull(types.StringType)
		}
	}
	if actionModel.Request.Custom.URL != "" {
		data.Request.Custom.URL = types.StringValue(actionModel.Request.Custom.URL)
	}
	if actionModel.Request.Custom.Body != "" {
		data.Request.Custom.Body = types.StringValue(actionModel.Request.Custom.Body)
	}
}
