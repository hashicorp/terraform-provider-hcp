// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		MarkdownDescription: "The Vault Secrets app resource manages an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Required ID field that is set to the app name.",
			},
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "The Vault Secrets App name.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`),
						"must contain only letters, numbers or hyphens",
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "The Vault Secrets app description",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the HCP Vault Secrets app is located.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the project the HCP Vault Secrets app is located.",
				Computed:    true,
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

type VaultSecretsApp struct {
	ID             types.String `tfsdk:"id"`
	AppName        types.String `tfsdk:"app_name"`
	Description    types.String `tfsdk:"description"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *resourceVaultsecretsApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VaultSecretsApp
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.CreateVaultSecretsApp(ctx, r.client, loc, plan.AppName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Vault Secrets App", err.Error())
		return
	}

	plan.ID = types.StringValue(res.Name)
	plan.AppName = types.StringValue(res.Name)
	plan.Description = types.StringValue(res.Description)
	plan.OrganizationID = types.StringValue(loc.OrganizationID)
	plan.ProjectID = types.StringValue(loc.ProjectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VaultSecretsApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.GetVaultSecretsApp(ctx, r.client, loc, state.AppName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to get app")
		return
	}

	state.AppName = types.StringValue(res.Name)
	state.Description = types.StringValue(res.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVaultsecretsApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VaultSecretsApp
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	res, err := clients.UpdateVaultSecretsApp(ctx, r.client, loc, plan.AppName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to get app")
		return
	}

	plan.ID = types.StringValue(res.Name)
	plan.AppName = types.StringValue(res.Name)
	plan.Description = types.StringValue(res.Description)
	plan.OrganizationID = types.StringValue(loc.OrganizationID)
	plan.ProjectID = types.StringValue(loc.ProjectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VaultSecretsApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      r.client.Config.ProjectID,
	}

	err := clients.DeleteVaultSecretsApp(ctx, r.client, loc, state.AppName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting app", err.Error())
		return
	}
}
