package testprovider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExampleDataSource struct {
	client *http.Client
}

type ExampleDataSourceModel struct {
	AppName   types.String `tfsdk:"app_name"`
	ProjectId types.String `tfsdk:"project_id"`
	OrgId     types.String `tfsdk:"organization_id"`
	Secret    types.String `tfsdk:"secret"`

	//Secrets   types.MapType `tfsdk:"secrets"`
}

func NewExampleDataSource() datasource.DataSource {
	return &ExampleDataSource{}
}

func (d *ExampleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vaultsecrets_app"
}

func (d *ExampleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Vault Secrets App Data Source",
		Attributes: map[string]schema.Attribute{
			"app_name": schema.StringAttribute{
				//TODO Add validator
				Description: "The name of the Vault Secrets application.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Vault Secrets app is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Vault Secrets app is located.",
				Computed:    true,
			},
			"secret": schema.StringAttribute{
				Description: "A test secret",
				Computed:    true,
			},
			/*"secrets": schema.MapAttribute{
				Description: "A map of all secrets in the Vault Secrets app. Key is the secret name, value is the latest secret version value.",
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},*/
		},
	}
}

func (d *ExampleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExampleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	//Add fake secrets
	data.Secret = types.StringValue("example-secret")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
