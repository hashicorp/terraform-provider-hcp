// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

var _ datasource.DataSource = &DataSourceAgentGroup{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceAgentGroup{}

func (d DataSourceAgentGroup) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("name"),
		),
	}
}

type DataSourceAgentGroup struct {
	client *clients.Client
}

func NewAgentGroupDataSource() datasource.DataSource {
	return &DataSourceAgentGroup{}
}

func (d *DataSourceAgentGroup) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_agent_group"
}

func (d *DataSourceAgentGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Agent Group resource manages the lifecycle of an Agent Group.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the Agent Group.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the Waypoint project to which the Agent Group belongs.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the Waypoint organization to which the Agent Group belongs.",
				Computed:    true,
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the Agent Group.",
				Optional:    true,
			},
		},
	}
}

func (d *DataSourceAgentGroup) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceAgentGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *AgentGroupResourceModel

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
	if !data.ProjectID.IsUnknown() && !data.ProjectID.IsNull() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := client.Config.OrganizationID
	if !data.OrgID.IsUnknown() && !data.OrgID.IsNull() {
		orgID = data.OrgID.ValueString()
	}

	group, err := clients.GetAgentGroup(ctx, client, &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}, data.Name.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			// If the group does not exist, remove it from state
			tflog.Info(ctx, "Waypoint Agent Group not found for organization, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Waypoint Agent Group", err.Error())
		return
	}

	if group.Description != "" {
		data.Description = types.StringValue(group.Description)
	} else {
		data.Description = types.StringNull()
	}

	if group.Name != "" {
		data.Name = types.StringValue(group.Name)
	} else {
		data.Name = types.StringNull()
	}

	if data.ProjectID.IsUnknown() {
		data.ProjectID = types.StringValue(projectID)
	}
	if data.OrgID.IsUnknown() {
		data.OrgID = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
