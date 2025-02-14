package vaultsecrets

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

var _ hvsResource = &Sync{}

var exactlyOneSyncConfigFieldsValidator = objectvalidator.ExactlyOneOf(
	path.Expressions{
		path.MatchRoot("gitlab_config"),
	}...,
)

type Sync struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	IntegrationName types.String `tfsdk:"integration_name"`
	ProjectID       types.String `tfsdk:"project_id"`
	OrganizationID  types.String `tfsdk:"organization_id"`

	// Destination-specific mutually exclusive fields
	GitlabConfig types.Object `tfsdk:"gitlab_config"`

	// Inner API-compatible models derived from the Terraform fields
	gitlabConfig *secretmodels.Secrets20231128SyncConfigGitlab `tfsdk:"-"`
}

func (s *Sync) projectID() types.String {
	return s.ProjectID
}

type gitlabConfigParams struct {
	Scope     types.String `tfsdk:"scope"`
	GroupID   types.String `tfsdk:"group_id"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (s *Sync) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)

	if !s.GitlabConfig.IsNull() {
		config := gitlabConfigParams{}
		diags := s.GitlabConfig.As(ctx, &config, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		scope := secretmodels.SyncConfigGitlabScope(config.Scope.ValueString())

		s.gitlabConfig = &secretmodels.Secrets20231128SyncConfigGitlab{
			GroupID:   config.GroupID.ValueString(),
			ProjectID: config.ProjectID.ValueString(),
			Protected: false,
			Raw:       false,
			Scope:     &scope,
		}
	}

	return diag.Diagnostics{}
}

func (s *Sync) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	syncModel, ok := model.(*secretmodels.Secrets20231128Sync)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128Sync, got: %T", model))
		return diags
	}

	s.Name = types.StringValue(syncModel.Name)
	s.IntegrationName = types.StringValue(syncModel.IntegrationName)
	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)
	s.ID = types.StringValue(syncModel.ResourceID)

	if syncModel.SyncConfigGitlab != nil {
		scope := *syncModel.SyncConfigGitlab.Scope
		var groupIDValue types.String
		var projectIDValue types.String

		if syncModel.SyncConfigGitlab.GroupID == "" {
			groupIDValue = types.StringNull()
		} else {
			groupIDValue = types.StringValue(syncModel.SyncConfigGitlab.GroupID)
		}

		if syncModel.SyncConfigGitlab.ProjectID == "" {
			projectIDValue = types.StringNull()
		} else {
			projectIDValue = types.StringValue(syncModel.SyncConfigGitlab.ProjectID)
		}

		s.GitlabConfig, diags = types.ObjectValue(
			s.GitlabConfig.AttributeTypes(ctx),
			map[string]attr.Value{
				"scope":      types.StringValue(string(scope)),
				"group_id":   groupIDValue,
				"project_id": projectIDValue,
			},
		)
	}

	return diags
}

var _ resource.Resource = &resourceVaultSecretsSync{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsSync{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsSync{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsSync{}

func NewVaultSecretsSyncResource() resource.Resource {
	return &resourceVaultSecretsSync{}
}

type resourceVaultSecretsSync struct {
	client *clients.Client
}

func (r *resourceVaultSecretsSync) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_sync"
}

func (r *resourceVaultSecretsSync) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Required ID field that is set to the sync name.",
		},
		"name": schema.StringAttribute{
			Description: "The Vault Secrets Sync name.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				slugValidator,
			},
		},
		"integration_name": schema.StringAttribute{
			Description: "The Vault Secrets integration name.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				slugValidator,
			},
		},
		"gitlab_config": schema.SingleNestedAttribute{
			Description: "Configuration parameters used to determine the sync destination.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"scope": schema.StringAttribute{
					Description: "The scope to which values apply. The valid options are GROUP and PROJECT",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("GROUP", "PROJECT"),
					},
				},
				"group_id": schema.StringAttribute{
					Description: "ID of the group, if the scope is GROUP",
					Optional:    true,
				},
				"project_id": schema.StringAttribute{
					Description: "ID of the project, if the scope is PROJECT",
					Optional:    true,
				},
			},
			Validators: []validator.Object{
				exactlyOneSyncConfigFieldsValidator,
			},
		},
	}

	maps.Copy(attributes, locationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets sync resource manages an integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsSync) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsSync) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsSync) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*Sync](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		sync, ok := i.(*Sync)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Sync, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetSync(secret_service.NewGetSyncParamsWithContext(ctx).
			WithOrganizationID(sync.OrganizationID.ValueString()).
			WithProjectID(sync.ProjectID.ValueString()).
			WithName(sync.Name.ValueString()), nil)

		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}

		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.Sync, nil
	})...)
}

func (r *resourceVaultSecretsSync) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*Sync](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		sync, ok := i.(*Sync)
		if !ok {
			return nil, fmt.Errorf("invalid resource type, expected *Sync, got: %T, this is a bug on the provider", i)
		}

		res, err := r.client.VaultSecrets.GetIntegration(secret_service.NewGetIntegrationParamsWithContext(ctx).
			WithOrganizationID(sync.OrganizationID.ValueString()).
			WithProjectID(sync.ProjectID.ValueString()).
			WithName(sync.IntegrationName.ValueString()), nil)

		if err != nil {
			return nil, err
		}

		providerType := res.Payload.Integration.Provider

		response, err := r.client.VaultSecrets.CreateSync(&secret_service.CreateSyncParams{
			Body: &secretmodels.SecretServiceCreateSyncBody{
				Name:             sync.Name.ValueString(),
				IntegrationName:  sync.IntegrationName.ValueString(),
				Type:             providerType,
				SyncConfigGitlab: sync.gitlabConfig,
			},
			OrganizationID: sync.OrganizationID.ValueString(),
			ProjectID:      sync.ProjectID.ValueString(),
		}, nil)
		if err != nil {
			return nil, err
		}

		return response.Payload.Sync, nil
	})...)
}

func (r *resourceVaultSecretsSync) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceVaultSecretsSync) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*Sync](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		sync, ok := i.(*Sync)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Sync, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecrets.DeleteSync(secret_service.NewDeleteSyncParamsWithContext(ctx).
			WithOrganizationID(sync.OrganizationID.ValueString()).
			WithProjectID(sync.ProjectID.ValueString()).
			WithName(sync.Name.ValueString()), nil)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsSync) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
