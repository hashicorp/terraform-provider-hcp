package provider

import (
	"context"
	"fmt"

	secrets "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-06-13/client/secret_service"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewVaultSecretsAppResource() *vaultsecretsAppResource {
	return &vaultsecretsAppResource{}
}

type vaultsecretsAppResource struct {
	client *clients.Client
}

func (r *vaultsecretsAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *vaultsecretsAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// TODO: Add validators
			"app_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *vaultsecretsAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}

// TODO Check that this is what gets returned
type App struct {
	AppName     string `tfsdk:"app_name"`
	Description string `tfsdk:"description"`
}

func (r *vaultsecretsAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data App
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	params := secrets.NewCreateAppParams()
	params.LocationOrganizationID = ""
	params.LocationProjectID = ""
	params.Body = secrets.NewCreateAppParams().Body
	params.Body.Name = data.AppName
	params.Body.Description = data.Description

	CreateVaultSecretsApp
	res, err := r.client.VaultSecrets.CreateApp(params, nil)
	if err != nil {
		//resp.Diagnostics.Append(err.Error(), "Unable to create app")
	}
	resp.State.Set(ctx, res.Payload.App)
}

func (r *vaultsecretsAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *vaultsecretsAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *vaultsecretsAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}
