package resourcemanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceOrganization struct {
	client *clients.Client
}

type DataSourceOrganizationModel struct {
	Name         types.String `tfsdk:"name"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceID   types.String `tfsdk:"resource_id"`
}

func NewOrganizationDataSource() datasource.DataSource {
	return &DataSourceOrganization{}
}

func (d *DataSourceOrganization) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *DataSourceOrganization) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The organization data source retrieves the HCP organization the provider is configured for.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The organization's name.",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: "The organization's resource name in format \"organization/<resource_id>\"",
				Computed:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The organization's unique identitier",
				Computed:    true,
			},
		},
	}
}

func (d *DataSourceOrganization) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceOrganization) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceOrganizationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Get the ID from the provider
	id := d.client.Config.OrganizationID

	getParams := organization_service.NewOrganizationServiceGetParams()
	getParams.ID = id
	res, err := d.client.Organization.OrganizationServiceGet(getParams, nil)
	if err != nil {
		var getErr *organization_service.OrganizationServiceGetDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Organization does not exist", fmt.Sprintf("unknown organization ID %q", id))
			return
		}

		resp.Diagnostics.AddError("Error retrieving organization", err.Error())
		return
	}

	o := res.GetPayload().Organization
	data.Name = types.StringValue(o.Name)
	data.ResourceName = types.StringValue(fmt.Sprintf("organization/%s", o.ID))
	data.ResourceID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
