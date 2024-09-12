// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
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

type IntegrationGCP struct {
	// Input fields
	ProjectID                 types.String `tfsdk:"project_id"`
	Name                      types.String `tfsdk:"name"`
	Capabilities              types.Set    `tfsdk:"capabilities"`
	ServiceAccountKey         types.Object `tfsdk:"service_account_key"`
	FederatedWorkloadIdentity types.Object `tfsdk:"federated_workload_identity"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`

	// Inner API-compatible models derived from the Terraform fields
	capabilities              []*secretmodels.Secrets20231128Capability                        `tfsdk:"-"`
	serviceAccountKey         *secretmodels.Secrets20231128GcpServiceAccountKeyRequest         `tfsdk:"-"`
	federatedWorkloadIdentity *secretmodels.Secrets20231128GcpFederatedWorkloadIdentityRequest `tfsdk:"-"`
}

// Helper structs to help populate concrete targets from types.Object fields
type serviceAccountKey struct {
	Credentials types.String `tfsdk:"credentials"`
	ProjectID   types.String `tfsdk:"project_id"`
	ClientEmail types.String `tfsdk:"client_email"`
}

type gcpFederatedWorkloadIdentity struct {
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
	Audience            types.String `tfsdk:"audience"`
}

var _ resource.Resource = &resourceVaultSecretsIntegrationGCP{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsIntegrationGCP{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsIntegrationGCP{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsIntegrationGCP{}

func NewVaultSecretsIntegrationGCPResource() resource.Resource {
	return &resourceVaultSecretsIntegrationGCP{}
}

type resourceVaultSecretsIntegrationGCP struct {
	client *clients.Client
}

func (r *resourceVaultSecretsIntegrationGCP) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_integration_gcp"
}

func (r *resourceVaultSecretsIntegrationGCP) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"service_account_key": schema.SingleNestedAttribute{
			Description: "GCP service account key used to authenticate against the target GCP project. Cannot be used with `federated_workload_identity`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"credentials": schema.StringAttribute{
					Description: "JSON or base64 encoded service account key received from GCP.",
					Required:    true,
				},
				"project_id": schema.StringAttribute{
					Description: "GCP project ID corresponding to the service account key.",
					Computed:    true,
				},
				"client_email": schema.StringAttribute{
					Description: "Service account email corresponding to the service account key.",
					Computed:    true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("federated_workload_identity"),
				}...),
			},
		},
		"federated_workload_identity": schema.SingleNestedAttribute{
			Description: "(Recommended) Federated identity configuration to authenticate against the target GCP project. Cannot be used with `service_account_key`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"service_account_email": schema.StringAttribute{
					Description: "GCP service account email that HVS will impersonate to carry operations for the appropriate capabilities.",
					Required:    true,
				},
				"audience": schema.StringAttribute{
					Description: "Audience configured on the GCP identity provider to federate access with HCP.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("service_account_key"),
				}...),
			},
		},
	}

	maps.Copy(attributes, sharedIntegrationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets GCP integration resource manages an GCP integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsIntegrationGCP) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsIntegrationGCP) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsIntegrationGCP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationGCP](ctx, r.client, &resp.State, req.State.Get, "reading", func(i integration) (any, error) {
		integration, ok := i.(*IntegrationGCP)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationGCP, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.GetGcpIntegration(
			secret_service.NewGetGcpIntegrationParamsWithContext(ctx).
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

func (r *resourceVaultSecretsIntegrationGCP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationGCP](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i integration) (any, error) {
		integration, ok := i.(*IntegrationGCP)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationGCP, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.CreateGcpIntegration(&secret_service.CreateGcpIntegrationParams{
			Body: &secretmodels.SecretServiceCreateGcpIntegrationBody{
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
				Name:                      integration.Name.ValueString(),
				ServiceAccountKey:         integration.serviceAccountKey,
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

func (r *resourceVaultSecretsIntegrationGCP) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationGCP](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i integration) (any, error) {
		integration, ok := i.(*IntegrationGCP)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationGCP, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.UpdateGcpIntegration(&secret_service.UpdateGcpIntegrationParams{
			Body: &secretmodels.SecretServiceUpdateGcpIntegrationBody{
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
				ServiceAccountKey:         integration.serviceAccountKey,
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

func (r *resourceVaultSecretsIntegrationGCP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationGCP](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i integration) (any, error) {
		integration, ok := i.(*IntegrationGCP)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationGCP, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecretsPreview.DeleteGcpIntegration(
			secret_service.NewDeleteGcpIntegrationParamsWithContext(ctx).
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationGCP) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Vault Secrets API does not return sensitive values like the service account key, so they will be initialized to an empty value
	// It means the first plan/apply after a successful import will always show a diff for the secret account key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

var _ integration = &IntegrationGCP{}

func (i *IntegrationGCP) projectID() types.String {
	return i.ProjectID
}

func (i *IntegrationGCP) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
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

	if !i.ServiceAccountKey.IsNull() {
		sa := serviceAccountKey{}
		diags = i.ServiceAccountKey.As(ctx, &sa, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.serviceAccountKey = &secretmodels.Secrets20231128GcpServiceAccountKeyRequest{
			Credentials: sa.Credentials.ValueString(),
		}
	}

	if !i.FederatedWorkloadIdentity.IsNull() {
		fwi := gcpFederatedWorkloadIdentity{}
		diags = i.FederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.federatedWorkloadIdentity = &secretmodels.Secrets20231128GcpFederatedWorkloadIdentityRequest{
			Audience:            fwi.Audience.ValueString(),
			ServiceAccountEmail: fwi.ServiceAccountEmail.ValueString(),
		}
	}

	return diag.Diagnostics{}
}

func (i *IntegrationGCP) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	integrationModel, ok := model.(*secretmodels.Secrets20231128GcpIntegration)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128GcpIntegration, got: %T", model))
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

	if integrationModel.ServiceAccountKey != nil {
		// The service account key is not returned by the API, so we use an empty value (e.g. for imports) or the state value (e.g. for updates)
		credentials := ""
		if i.serviceAccountKey != nil {
			credentials = i.serviceAccountKey.Credentials
		}

		i.ServiceAccountKey, diags = types.ObjectValue(i.ServiceAccountKey.AttributeTypes(ctx), map[string]attr.Value{
			"credentials":  types.StringValue(credentials),
			"project_id":   types.StringValue(integrationModel.ServiceAccountKey.ProjectID),
			"client_email": types.StringValue(integrationModel.ServiceAccountKey.ClientEmail),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.FederatedWorkloadIdentity != nil {
		i.FederatedWorkloadIdentity, diags = types.ObjectValue(i.FederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"service_account_email": types.StringValue(integrationModel.FederatedWorkloadIdentity.ServiceAccountEmail),
			"audience":              types.StringValue(integrationModel.FederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
