// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceProject struct {
	client *clients.Client
}

type DataSourceProjectModel struct {
	Project      types.String `tfsdk:"project"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceID   types.String `tfsdk:"resource_id"`
}

func NewProjectDataSource() datasource.DataSource {
	return &DataSourceProject{}
}

func (d *DataSourceProject) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *DataSourceProject) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The project data source retrieves the given HCP project.",
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Description: "The id of the project. May be given as \"<id>\" or \"project/<id>\". If not set, the provider project is used.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The project's name.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The project's description",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: "The project's resource name in format \"project/<resource_id>\"",
				Computed:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The project's unique identifier",
				Computed:    true,
			},
		},
	}
}

func (d *DataSourceProject) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceProjectModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Determine the ID
	id := data.Project.ValueString()
	id = strings.TrimPrefix(id, "project/")
	if id == "" {
		id = d.client.Config.ProjectID
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	getParams := project_service.NewProjectServiceGetParams()
	getParams.ID = id
	res, err := d.client.Project.ProjectServiceGet(getParams, nil)
	if err != nil {
		var getErr *project_service.ProjectServiceGetDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Project does not exist", fmt.Sprintf("unknown project ID %q", id))
			return
		}

		resp.Diagnostics.AddError("Error retrieving project", err.Error())
		return
	}

	p := res.GetPayload().Project
	data.Description = types.StringValue(p.Description)
	data.Name = types.StringValue(p.Name)
	data.ResourceName = types.StringValue(fmt.Sprintf("project/%s", p.ID))
	data.ResourceID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
