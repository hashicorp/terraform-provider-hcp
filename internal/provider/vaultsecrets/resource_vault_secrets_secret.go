// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"
	"regexp"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

var _ resource.Resource = &resourceVaultsecretsSecret{}
var _ resource.ResourceWithConfigure = &resourceVaultsecretsSecret{}
var _ resource.ResourceWithModifyPlan = &resourceVaultsecretsSecret{}

func NewVaultSecretsSecretResource() resource.Resource {
	return &resourceVaultsecretsSecret{}
}

type resourceVaultsecretsSecret struct {
	client *clients.Client
}

type VaultSecretsSecret struct {
	ID             types.String `tfsdk:"id"`
	AppName        types.String `tfsdk:"app_name"`
	SecretName     types.String `tfsdk:"secret_name"`
	SecretValue    types.String `tfsdk:"secret_value"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *resourceVaultsecretsSecret) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_secret"
}

func (r *resourceVaultsecretsSecret) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets secret resource manages a secret within a given application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the resource",
				Computed:    true,
			},
			"app_name": schema.StringAttribute{
				Description: "The name of the application the secret can be found in",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(36),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]+[a-zA-Z0-9]$`),
						"must contain only ASCII letters, numbers, and hyphens; must not start or end with a hyphen",
					),
				},
			},
			"secret_name": schema.StringAttribute{
				Description: "The name of the secret",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
						"must contain only ASCII letters, numbers, and underscores; must not start with a number",
					),
				},
			},
			"secret_value": schema.StringAttribute{
				Description: "The value of the secret",
				Required:    true,
				Sensitive:   true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the HCP Vault Secrets secret is located.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the project the HCP Vault Secrets secret is located.",
				Computed:    true,
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

func (r *resourceVaultsecretsSecret) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultsecretsSecret) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VaultSecretsSecret
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	res, err := clients.CreateVaultSecretsAppSecret(ctx, r.client, loc, plan.AppName.ValueString(), plan.SecretName.ValueString(), plan.SecretValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret", err.Error())
		return
	}

	plan.ID = plan.AppName
	plan.SecretName = types.StringValue(res.Name)
	plan.OrganizationID = types.StringValue(loc.OrganizationID)
	plan.ProjectID = types.StringValue(loc.ProjectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsSecret) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VaultSecretsSecret
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	res, err := clients.OpenVaultSecretsAppSecret(ctx, r.client, loc, state.AppName.ValueString(), state.SecretName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Error reading secret")
		return
	}

	// TODO: so the resource can only create a static secret,
	// what happens when a user tries to import a rotating/other type of secret?
	state.SecretValue = types.StringValue(res.StaticVersion.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVaultsecretsSecret) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VaultSecretsSecret
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	res, err := clients.CreateVaultSecretsAppSecret(ctx, r.client, loc, plan.AppName.ValueString(), plan.SecretName.ValueString(), plan.SecretValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}

	plan.ID = plan.AppName
	plan.SecretName = types.StringValue(res.Name)
	plan.OrganizationID = types.StringValue(loc.OrganizationID)
	plan.ProjectID = types.StringValue(loc.ProjectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVaultsecretsSecret) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VaultSecretsSecret
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !state.ProjectID.IsUnknown() {
		projectID = state.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	err := clients.DeleteVaultSecretsAppSecret(ctx, r.client, loc, state.AppName.ValueString(), state.SecretName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
		return
	}
}
