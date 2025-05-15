// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcemanager

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceProjects struct {
	client *clients.Client
}

type ProjectModel struct {
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceID   types.String `tfsdk:"resource_id"`
}

type DataSourceProjectsModel struct {
	Projects []ProjectModel `tfsdk:"projects"`
}

func NewProjectsDataSource() datasource.DataSource {
	return &DataSourceProjects{}
}

func (d *DataSourceProjects) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *DataSourceProjects) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The projects data source retrieves all projects in the given HCP organization.",
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Description: "A list of projects in the HCP organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": &schema.StringAttribute{
							Description: "The project's name.",
							Computed:    true,
						},
						"description": &schema.StringAttribute{
							Description: "The project's description.",
							Computed:    true,
							Optional:    true,
						},
						"resource_name": &schema.StringAttribute{
							Description: "The project's resource name.",
							Computed:    true,
						},
						"resource_id": &schema.StringAttribute{
							Description: "The project's unique identifier",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DataSourceProjects) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceProjects) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceProjectsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	listParams := project_service.NewProjectServiceListParamsWithContext(ctx)
	listParams.ScopeID = &d.client.Config.OrganizationID
	scopeType := string(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	listParams.ScopeType = &scopeType
	res, err := d.client.Project.ProjectServiceList(listParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving projects", err.Error())
		return
	}

	p := res.GetPayload().Projects
	projects := make([]ProjectModel, len(p))
	for i, project := range p {
		projects[i] = ProjectModel{
			Name:         types.StringValue(project.Name),
			Description:  types.StringValue(project.Description),
			ResourceName: types.StringValue(fmt.Sprintf("project/%s", project.ID)),
			ResourceID:   types.StringValue(project.ID),
		}
	}

	data.Projects = projects
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
