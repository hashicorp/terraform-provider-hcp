// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

type App struct {
	ID             types.String `tfsdk:"id"`
	AppName        types.String `tfsdk:"app_name"`
	Description    types.String `tfsdk:"description"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceName   types.String `tfsdk:"resource_name"`
	SyncNames      types.Set    `tfsdk:"sync_names"`

	syncNames []string `tfsdk:"-"`
}

var _ resource.Resource = &resourceVaultSecretsApp{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsApp{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsApp{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsApp{}

func NewVaultSecretsAppResource() resource.Resource {
	return &resourceVaultSecretsApp{}
}

type resourceVaultSecretsApp struct {
	client *clients.Client
}

func (r *resourceVaultSecretsApp) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_app"
}

func (r *resourceVaultSecretsApp) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the project the HCP Vault Secrets app is located.",
				Computed:    true,
			},
			"resource_name": schema.StringAttribute{
				Computed:    true,
				Description: "The app's resource name in the format secrets/project/<project ID>/app/<app Name>.",
			},
			"sync_names": schema.SetAttribute{
				Description: "Set of sync names to associate with this app.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						slugValidator,
					),
				},
			}},
	}
}

func (r *resourceVaultSecretsApp) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsApp) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*App](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		app, ok := i.(*App)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *App, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.CreateApp(&secret_service.CreateAppParams{
			Body: &secretmodels.SecretServiceCreateAppBody{
				Name:        app.AppName.ValueString(),
				Description: app.Description.ValueString(),
				SyncNames:   app.syncNames,
			},
			OrganizationID: app.OrganizationID.ValueString(),
			ProjectID:      app.ProjectID.ValueString(),
		}, nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.App, nil
	})...)
}

func (r *resourceVaultSecretsApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*App](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		app, ok := i.(*App)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *App, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetApp(
			secret_service.NewGetAppParamsWithContext(ctx).
				WithOrganizationID(app.OrganizationID.ValueString()).
				WithProjectID(app.ProjectID.ValueString()).
				WithName(app.AppName.ValueString()), nil)
		if err != nil && !clients.IsResponseForbidden(err) { // The HVS API returns 403 if the app doesn't exist even if the principal has the correct permissions.
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.App, nil
	})...)
}

func (r *resourceVaultSecretsApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*App](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		app, ok := i.(*App)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *App, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.UpdateApp(&secret_service.UpdateAppParams{
			Body: &secretmodels.SecretServiceUpdateAppBody{
				Description: app.Description.ValueString(),
			},
			Name:           app.AppName.ValueString(),
			OrganizationID: app.OrganizationID.ValueString(),
			ProjectID:      app.ProjectID.ValueString(),
		}, nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.App, nil
	})...)
}

func (r *resourceVaultSecretsApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*App](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		app, ok := i.(*App)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *App, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecrets.DeleteApp(
			secret_service.NewDeleteAppParamsWithContext(ctx).
				WithOrganizationID(app.OrganizationID.ValueString()).
				WithProjectID(app.ProjectID.ValueString()).
				WithName(app.AppName.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsApp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_name"), req.ID)...)
}

var _ hvsResource = &App{}

func (a *App) projectID() types.String {
	return a.ProjectID
}

func (a *App) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
	a.OrganizationID = types.StringValue(orgID)
	a.ProjectID = types.StringValue(projID)
	a.syncNames = make([]string, 0, len(a.SyncNames.Elements()))
	a.SyncNames.ElementsAs(ctx, &a.syncNames, false)

	return diag.Diagnostics{}
}

func (a *App) fromModel(_ context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	appModel, ok := model.(*secretmodels.Secrets20231128App)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128App, got: %T", model))
		return diags
	}

	a.OrganizationID = types.StringValue(orgID)
	a.ProjectID = types.StringValue(projID)
	a.ID = types.StringValue(appModel.ResourceID)
	a.ResourceName = types.StringValue(appModel.ResourceName)

	var syncs []attr.Value
	for _, c := range appModel.SyncNames {
		syncs = append(syncs, types.StringValue(c))
	}
	a.SyncNames, diags = types.SetValue(types.StringType, syncs)
	if diags.HasError() {
		return diags
	}

	return diags
}
