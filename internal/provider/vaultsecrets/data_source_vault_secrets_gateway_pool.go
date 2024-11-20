package vaultsecrets

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceVaultSecretsGatewayPool struct {
	client *clients.Client
}

type DataSourceVaultSecretsGatewayPoolModel struct {
	GatewayPoolName types.String `tfsdk:"gateway_pool_name"`
	OrgID           types.String `tfsdk:"organization_id"`
	ProjectID       types.String `tfsdk:"project_id"`
}

func NewVaultSecretsGatewayPoolDataSource() datasource.DataSource {
	return &DataSourceVaultSecretsGatewayPool{}
}

func (d *DataSourceVaultSecretsGatewayPool) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_gateway_pool"
}

func (d *DataSourceVaultSecretsGatewayPool) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (d *DataSourceVaultSecretsGatewayPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}
