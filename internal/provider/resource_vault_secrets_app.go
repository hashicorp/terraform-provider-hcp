package provider

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clients "github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewVaultSecretsAppResource() resource.Resource {
	return &resourceVaultsecretsApp{}
}

type resourceVaultsecretsApp struct {
	client *clients.Client
}

func (r *resourceVaultsecretsApp) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_app"
}

func (r *resourceVaultsecretsApp) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
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

func (r *resourceVaultsecretsApp) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type App struct {
	ID          types.String `tfsdk:"id"`
	AppName     types.String `tfsdk:"app_name"`
	Description types.String `tfsdk:"description"`
}

func (r *resourceVaultsecretsApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan App
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.CreateVaultSecretsApp(ctx, r.client, loc, plan.AppName.String(), plan.Description.String())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Vault Secrets App", err.Error())
		return
	}

	plan.ID = types.StringValue(res.Name)
	plan.AppName = types.StringValue(res.Name)
	plan.Description = types.StringValue(res.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state App
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.GetVaultSecretsApp(ctx, r.client, loc, state.AppName.String())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to get app")
	}

	state.AppName = types.StringValue(res.Name)
	state.Description = types.StringValue(res.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVaultsecretsApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan App
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.UpdateVaultSecretsApp(ctx, r.client, loc, plan.AppName.String(), plan.Description.String())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to get app")
	}

	plan.AppName = types.StringValue(res.Name)
	plan.Description = types.StringValue(res.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state App
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	err := clients.DeleteVaultSecretsApp(ctx, r.client, loc, state.AppName.String())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting app", err.Error())
		return
	}
}
