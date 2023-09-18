package provider

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

func NewVaultSecretsSecretResource() resource.Resource {
	return &resourceVaultsecretsSecret{}
}

type resourceVaultsecretsSecret struct {
	client *clients.Client
}

type Secret struct {
	ID          types.String `tfsdk:"id"`
	AppName     string       `tfsdk:"app_name"`
	SecretName  string       `tfsdk:"secret_name"`
	SecretValue string       `tfsdk:"secret_value"`
}

func (r *resourceVaultsecretsSecret) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_secret"
}

func (r *resourceVaultsecretsSecret) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			// TODO: Add validators
			"app_name": schema.StringAttribute{
				Required: true,
			},
			"secret_name": schema.StringAttribute{
				Required: true,
			},
			"secret_value": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *resourceVaultsecretsSecret) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Source Configure Type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *resourceVaultsecretsSecret) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Secret
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.CreateVaultSecretsAppSecret(ctx, r.client, loc, plan.AppName, plan.SecretName, plan.SecretValue)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Vault Secrets Secret", err.Error())
		return
	}

	// TODO: add more to plan here?
	plan.ID = types.StringValue(plan.AppName)
	plan.SecretName = res.Name
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceVaultsecretsSecret) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Secret
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.OpenVaultSecretsAppSecret(ctx, r.client, loc, state.AppName, state.SecretName)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to get secret")
	}

	state.SecretName = res.Name
	state.SecretValue = res.LatestVersion

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceVaultsecretsSecret) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Secret
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.CreateVaultSecretsAppSecret(ctx, r.client, loc, plan.AppName, plan.SecretName, plan.SecretValue)

	if err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}

	plan.SecretName = res.Name
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *resourceVaultsecretsSecret) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Secret
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	err := clients.DeleteVaultSecretsAppSecret(ctx, r.client, loc, state.AppName, state.SecretName)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
		return
	}
}
