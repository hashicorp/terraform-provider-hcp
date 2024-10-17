// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
	"golang.org/x/exp/maps"
)

type IntegrationConfluent struct {
	// Input fields
	ProjectID               types.String `tfsdk:"project_id"`
	Name                    types.String `tfsdk:"name"`
	Capabilities            types.Set    `tfsdk:"capabilities"`
	StaticCredentialDetails types.Object `tfsdk:"static_credential_details"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`

	// Inner API-compatible models derived from the Terraform fields
	capabilities            []*secretmodels.Secrets20231128Capability                      `tfsdk:"-"`
	staticCredentialDetails *secretmodels.Secrets20231128ConfluentStaticCredentialsRequest `tfsdk:"-"`
}

// Helper structs to help populate concrete targets from types.Object fields
type confluentStaticCredentialDetails struct {
	CloudAPIKeyID  types.String `tfsdk:"cloud_api_key_id"`
	CloudAPISecret types.String `tfsdk:"cloud_api_secret"`
}

var _ resource.Resource = &resourceVaultSecretsIntegrationConfluent{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsIntegrationConfluent{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsIntegrationConfluent{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsIntegrationConfluent{}

func NewVaultSecretsIntegrationsConfluentResource() resource.Resource {
	return &resourceVaultSecretsIntegrationConfluent{}
}

type resourceVaultSecretsIntegrationConfluent struct {
	client *clients.Client
}

func (r *resourceVaultSecretsIntegrationConfluent) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_integration_Confluent"
}

func (r *resourceVaultSecretsIntegrationConfluent) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"static_credential_details": schema.SingleNestedAttribute{
			Description: "Confluent API key used to authenticate against the target project.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"cloud_api_key_id": schema.StringAttribute{
					Description: "Public key used alongside the private key to authenticate against the target project.",
					Required:    true,
				},
				"cloud_api_secret": schema.StringAttribute{
					Description: "Private key used alongside the public key to authenticate against the target project.",
					Required:    true,
					Sensitive:   true,
				},
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, sharedIntegrationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets Confluent integration resource manages an Confluent integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsIntegrationConfluent) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsIntegrationConfluent) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsIntegrationConfluent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationConfluent](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationConfluent)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationConfluent, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.GetConfluentIntegration(
			secret_service.NewGetConfluentIntegrationParamsWithContext(ctx).
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

func (r *resourceVaultSecretsIntegrationConfluent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationConfluent](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationConfluent)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationConfluent, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.CreateConfluentIntegration(&secret_service.CreateConfluentIntegrationParams{
			Body: &secretmodels.SecretServiceCreateConfluentIntegrationBody{
				Capabilities:            integration.capabilities,
				Name:                    integration.Name.ValueString(),
				StaticCredentialDetails: integration.staticCredentialDetails,
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

func (r *resourceVaultSecretsIntegrationConfluent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationConfluent](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationConfluent)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationConfluent, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecretsPreview.UpdateConfluentIntegration(&secret_service.UpdateConfluentIntegrationParams{
			Body: &secretmodels.SecretServiceUpdateConfluentIntegrationBody{
				Capabilities:            integration.capabilities,
				StaticCredentialDetails: integration.staticCredentialDetails,
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

func (r *resourceVaultSecretsIntegrationConfluent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationConfluent](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationConfluent)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationConfluent, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecretsPreview.DeleteConfluentIntegration(
			secret_service.NewDeleteConfluentIntegrationParamsWithContext(ctx).
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationConfluent) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Vault Secrets API does not return sensitive values like the secret access key, so they will be initialized to an empty value
	// It means the first plan/apply after a successful import will always show a diff for the secret access key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

var _ hvsResource = &IntegrationConfluent{}

func (i *IntegrationConfluent) projectID() types.String {
	return i.ProjectID
}

func (i *IntegrationConfluent) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
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

	if !i.StaticCredentialDetails.IsNull() {
		scd := confluentStaticCredentialDetails{}
		diags = i.StaticCredentialDetails.As(ctx, &scd, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.staticCredentialDetails = &secretmodels.Secrets20231128ConfluentStaticCredentialsRequest{
			CloudAPIKeyID:  scd.CloudAPIKeyID.ValueString(),
			CloudAPISecret: scd.CloudAPISecret.ValueString(),
		}
	}

	return diag.Diagnostics{}
}

func (i *IntegrationConfluent) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	integrationModel, ok := model.(*secretmodels.Secrets20231128ConfluentIntegration)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128ConfluentIntegration, got: %T", model))
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

	if integrationModel.StaticCredentialDetails != nil {
		// The secret key is not returned by the API, so we use an empty value (e.g. for imports) or the state value (e.g. for updates)
		cloudApiSecret := ""
		if i.staticCredentialDetails != nil {
			cloudApiSecret = i.staticCredentialDetails.CloudAPISecret
		}

		i.StaticCredentialDetails, diags = types.ObjectValue(i.StaticCredentialDetails.AttributeTypes(ctx), map[string]attr.Value{
			"cloud_api_key_id": types.StringValue(integrationModel.StaticCredentialDetails.CloudAPIKeyID),
			"cloud_api_secret": types.StringValue(cloudApiSecret),
		})
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
