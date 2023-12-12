// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serviceprincipal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewDataSource() datasource.DataSource {
	return &dataSourceConfig{}
}

type dataSourceConfig struct {
	client *clients.Client
}

var _ datasource.DataSource = &dataSourceConfig{}

type dataSourceModel struct {
	Name         types.String `tfsdk:"name"`
	ResourceName types.String `tfsdk:"resource_name"`
	ResourceID   types.String `tfsdk:"resource_id"`
}

func (d *dataSourceConfig) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_principal"
}

func (d *dataSourceConfig) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The service principal data source retrieves the given HCP Service Principal.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The service principal's name",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Description: "The service principal's resource name in the format `iam/project/<project_id>/service-principal/<name>` or `iam/organization/<organization_id>/service-principal/<name>`",
				Required:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The service principal's unique identitier",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceConfig) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataSourceConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	getParams := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParams()
	getParams.ResourceName = data.ResourceName.ValueString()
	res, err := d.client.ServicePrincipals.ServicePrincipalsServiceGetServicePrincipal(getParams, nil)
	if err != nil {
		var getErr *service_principals_service.ServicePrincipalsServiceGetServicePrincipalDefault
		if errors.As(err, &getErr) && getErr.IsCode(http.StatusNotFound) {
			resp.Diagnostics.AddError("Service principal does not exist", fmt.Sprintf("unknown service principal %q", data.ResourceName.ValueString()))
			return
		}

		resp.Diagnostics.AddError("Error retrieving service principal", err.Error())
		return
	}

	sp := res.GetPayload().ServicePrincipal
	data.Name = types.StringValue(sp.Name)
	data.ResourceName = types.StringValue(sp.ResourceName)
	data.ResourceID = types.StringValue(sp.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
