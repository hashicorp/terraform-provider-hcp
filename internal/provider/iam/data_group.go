package iam

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceGroup struct {
	client *clients.Client
}

type DataSourceGroupModel struct {
	DisplayName  types.String `tfsdk:"display_name"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceID   types.String `tfsdk:"resource_id"`
	Description  types.String `tfsdk:"description"`
}

func NewGroupDataSource() datasource.DataSource {
	return &DataSourceGroup{}
}

func (d *DataSourceGroup) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *DataSourceGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The group data source retrieves the given group.",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Description: "The group's display name",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: fmt.Sprintf("The group's resource name in format `%s` or shortened `%s`",
					"iam/organization/<organization_id>/group/<resource_name>", "<resource_name>"),
				Required: true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The group's unique identifier",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The group's description",
				Computed:    true,
			},
		},
	}
}

func (d *DataSourceGroup) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	getParams := groups_service.NewGroupsServiceGetGroupParams()
	getParams.ResourceName = data.ResourceName.ValueString()

	// if shorthand resourceName was provided, generate full resourceName
	if !strings.HasPrefix(getParams.ResourceName, "iam/") {
		orgID := d.client.Config.OrganizationID
		fmt.Printf("orgID is : %s", orgID)
		getParams.ResourceName = fmt.Sprintf("iam/organization/%s/group/%s", orgID, data.ResourceName.ValueString())
	}

	res, err := d.client.Groups.GroupsServiceGetGroup(getParams, nil)

	if err != nil {
		var getErr *groups_service.GroupsServiceGetGroupDefault

		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Group does not exist", fmt.Sprintf("unknown group %q", data.ResourceName.ValueString()))
			return
		}

		resp.Diagnostics.AddError("Error retrieving group", err.Error())
		return
	}

	group := res.GetPayload().Group
	data.DisplayName = types.StringValue(group.DisplayName)
	data.ResourceName = types.StringValue(group.ResourceName)
	data.ResourceID = types.StringValue(group.ResourceID)
	data.Description = types.StringValue(group.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
