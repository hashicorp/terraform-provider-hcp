// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
	"golang.org/x/exp/maps"
)

type IntegrationAzure struct {
	// Input fields
	ProjectID                 types.String `tfsdk:"project_id"`
	Name                      types.String `tfsdk:"name"`
	Capabilities              types.Set    `tfsdk:"capabilities"`
	ClientSecret              types.Object `tfsdk:"client_secret"`
	FederatedWorkloadIdentity types.Object `tfsdk:"federated_workload_identity"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`

	// Inner API-compatible models derived from the Terraform fields
	capabilities              []*secretmodels.Secrets20231128Capability                          `tfsdk:"-"`
	clientSecret              *secretmodels.Secrets20231128AzureClientSecretRequest              `tfsdk:"-"`
	federatedWorkloadIdentity *secretmodels.Secrets20231128AzureFederatedWorkloadIdentityRequest `tfsdk:"-"`
}

// Helper structs to help populate concrete targets from types.Object fields
type clientSecret struct {
	TenantID     types.String `tfsdk:"tenant_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type azureFederatedWorkloadIdentity struct {
	TenantID types.String `tfsdk:"tenant_id"`
	ClientID types.String `tfsdk:"client_id"`
	Audience types.String `tfsdk:"audience"`
}

var _ resource.Resource = &resourceVaultSecretsIntegrationAzure{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsIntegrationAzure{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsIntegrationAzure{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsIntegrationAzure{}

func NewVaultSecretsIntegrationAzureResource() resource.Resource {
	return &resourceVaultSecretsIntegrationAzure{}
}

type resourceVaultSecretsIntegrationAzure struct {
	client *clients.Client
}

func (r *resourceVaultSecretsIntegrationAzure) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_integration_azure"
}

func (r *resourceVaultSecretsIntegrationAzure) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"client_secret": schema.SingleNestedAttribute{
			Description: "Azure client secret used to authenticate against the target Azure application. Cannot be used with `federated_workload_identity`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"tenant_id": schema.StringAttribute{
					Description: "Azure tenant ID corresponding to the Azure application.",
					Required:    true,
				},
				"client_id": schema.StringAttribute{
					Description: "Azure client ID corresponding to the Azure application.",
					Required:    true,
				},
				"client_secret": schema.StringAttribute{
					Description: "Secret value corresponding to the Azure client secret.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("federated_workload_identity"),
				}...),
			},
		},
		"federated_workload_identity": schema.SingleNestedAttribute{
			Description: "(Recommended) Federated identity configuration to authenticate against the target Azure application. Cannot be used with `client_secret`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"tenant_id": schema.StringAttribute{
					Description: "Azure tenant ID corresponding to the Azure application.",
					Required:    true,
				},
				"client_id": schema.StringAttribute{
					Description: "Azure client ID corresponding to the Azure application.",
					Required:    true,
				},
				"audience": schema.StringAttribute{
					Description: "Audience configured on the Azure federated identity credentials to federate access with HCP.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("client_secret"),
				}...),
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, sharedIntegrationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets Azure integration resource manages an Azure integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsIntegrationAzure) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsIntegrationAzure) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsIntegrationAzure) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAzure](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAzure)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAzure, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetAzureIntegration(
			secret_service.NewGetAzureIntegrationParamsWithContext(ctx).
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.Integration, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationAzure) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAzure](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAzure)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAzure, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.CreateAzureIntegration(&secret_service.CreateAzureIntegrationParams{
			Body: &secretmodels.SecretServiceCreateAzureIntegrationBody{
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
				Name:                      integration.Name.ValueString(),
				ClientSecret:              integration.clientSecret,
			},
			OrganizationID: integration.OrganizationID.ValueString(),
			ProjectID:      integration.ProjectID.ValueString(),
		}, nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.Integration, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationAzure) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAzure](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAzure)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAzure, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.UpdateAzureIntegration(&secret_service.UpdateAzureIntegrationParams{
			Body: &secretmodels.SecretServiceUpdateAzureIntegrationBody{
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
				ClientSecret:              integration.clientSecret,
			},
			Name:           integration.Name.ValueString(),
			OrganizationID: integration.OrganizationID.ValueString(),
			ProjectID:      integration.ProjectID.ValueString(),
		}, nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		if response == nil || response.Payload == nil {
			return nil, nil
		}
		return response.Payload.Integration, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationAzure) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAzure](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAzure)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAzure, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecrets.DeleteAzureIntegration(
			secret_service.NewDeleteAzureIntegrationParams().
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationAzure) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Vault Secrets API does not return sensitive values like the client secret, so they will be initialized to an empty value
	// It means the first plan/apply after a successful import will always show a diff for the client secret.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

var _ hvsResource = &IntegrationAzure{}

func (i *IntegrationAzure) projectID() types.String {
	return i.ProjectID
}

func (i *IntegrationAzure) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
	// Init fields that depend on the Terraform provider configuration
	i.OrganizationID = types.StringValue(orgID)
	i.ProjectID = types.StringValue(projID)

	// Init the HVS domain models from the Terraform domain models
	var capabilities []types.String
	diags := i.Capabilities.ElementsAs(ctx, &capabilities, false)
	if diags.HasError() {
		return diags
	}
	for _, c := range capabilities {
		i.capabilities = append(i.capabilities, secretmodels.Secrets20231128Capability(c.ValueString()).Pointer())
	}

	if !i.ClientSecret.IsNull() {
		cs := clientSecret{}
		diags = i.ClientSecret.As(ctx, &cs, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		// The service account key is not returned by the API, so we use an empty value (e.g. for imports) or the state value (e.g. for updates)
		clientSecret := ""
		if i.clientSecret != nil {
			clientSecret = i.clientSecret.ClientSecret
		}

		i.clientSecret = &secretmodels.Secrets20231128AzureClientSecretRequest{
			TenantID:     clientSecret,
			ClientID:     cs.ClientID.ValueString(),
			ClientSecret: cs.ClientSecret.ValueString(),
		}
	}

	if !i.FederatedWorkloadIdentity.IsNull() {
		fwi := azureFederatedWorkloadIdentity{}
		diags = i.FederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.federatedWorkloadIdentity = &secretmodels.Secrets20231128AzureFederatedWorkloadIdentityRequest{
			Audience: fwi.Audience.ValueString(),
			TenantID: fwi.TenantID.ValueString(),
			ClientID: fwi.ClientID.ValueString(),
		}
	}

	return diag.Diagnostics{}
}

func (i *IntegrationAzure) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	integrationModel, ok := model.(*secretmodels.Secrets20231128AzureIntegration)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128AzureIntegration, got: %T", model))
		return diags
	}

	i.OrganizationID = types.StringValue(orgID)
	i.ProjectID = types.StringValue(projID)
	i.ResourceID = types.StringValue(integrationModel.ResourceID)
	i.ResourceName = types.StringValue(integrationModel.ResourceName)
	i.Name = types.StringValue(integrationModel.Name)

	var values []attr.Value
	for _, c := range integrationModel.Capabilities {
		values = append(values, types.StringValue(string(*c)))
	}
	i.Capabilities, diags = types.SetValue(types.StringType, values)
	if diags.HasError() {
		return diags
	}

	if integrationModel.ClientSecret != nil {
		i.ClientSecret, diags = types.ObjectValue(i.ClientSecret.AttributeTypes(ctx), map[string]attr.Value{
			"tenant_id": types.StringValue(integrationModel.ClientSecret.TenantID),
			"client_id": types.StringValue(integrationModel.ClientSecret.ClientID),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.FederatedWorkloadIdentity != nil {
		i.FederatedWorkloadIdentity, diags = types.ObjectValue(i.FederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"tenant_id": types.StringValue(integrationModel.FederatedWorkloadIdentity.TenantID),
			"client_id": types.StringValue(integrationModel.FederatedWorkloadIdentity.ClientID),
			"audience":  types.StringValue(integrationModel.FederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
