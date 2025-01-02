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

type IntegrationAWS struct {
	// Input fields
	ProjectID                 types.String `tfsdk:"project_id"`
	Name                      types.String `tfsdk:"name"`
	Capabilities              types.Set    `tfsdk:"capabilities"`
	AccessKeys                types.Object `tfsdk:"access_keys"`
	FederatedWorkloadIdentity types.Object `tfsdk:"federated_workload_identity"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`

	// Inner API-compatible models derived from the Terraform fields
	capabilities              []*secretmodels.Secrets20231128Capability                        `tfsdk:"-"`
	accessKeys                *secretmodels.Secrets20231128AwsAccessKeysRequest                `tfsdk:"-"`
	federatedWorkloadIdentity *secretmodels.Secrets20231128AwsFederatedWorkloadIdentityRequest `tfsdk:"-"`
}

// Helper structs to help populate concrete targets from types.Object fields
type accessKeys struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type awsFederatedWorkloadIdentity struct {
	RoleARN  types.String `tfsdk:"role_arn"`
	Audience types.String `tfsdk:"audience"`
}

var _ resource.Resource = &resourceVaultSecretsIntegrationAWS{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsIntegrationAWS{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsIntegrationAWS{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsIntegrationAWS{}

func NewVaultSecretsIntegrationAWSResource() resource.Resource {
	return &resourceVaultSecretsIntegrationAWS{}
}

type resourceVaultSecretsIntegrationAWS struct {
	client *clients.Client
}

func (r *resourceVaultSecretsIntegrationAWS) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_integration_aws"
}

func (r *resourceVaultSecretsIntegrationAWS) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"access_keys": schema.SingleNestedAttribute{
			Description: "AWS IAM key pair used to authenticate against the target AWS account. Cannot be used with `federated_workload_identity`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"access_key_id": schema.StringAttribute{
					Description: "Key ID used with the secret key to authenticate against the target AWS account.",
					Required:    true,
				},
				"secret_access_key": schema.StringAttribute{
					Description: "Secret key used with the key ID to authenticate against the target AWS account.",
					Required:    true,
					Sensitive:   true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("federated_workload_identity"),
				}...),
			},
		},
		"federated_workload_identity": schema.SingleNestedAttribute{
			Description: "(Recommended) Federated identity configuration to authenticate against the target AWS account. Cannot be used with `access_keys`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"role_arn": schema.StringAttribute{
					Description: "AWS IAM role ARN the integration will assume to carry operations for the appropriate capabilities.",
					Required:    true,
				},
				"audience": schema.StringAttribute{
					Description: "Audience configured on the AWS IAM identity provider to federate access with HCP.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRoot("access_keys"),
				}...),
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, sharedIntegrationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets AWS integration resource manages an AWS integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsIntegrationAWS) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsIntegrationAWS) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsIntegrationAWS) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAWS](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAWS)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAWS, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetAwsIntegration(
			secret_service.NewGetAwsIntegrationParamsWithContext(ctx).
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

func (r *resourceVaultSecretsIntegrationAWS) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAWS](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAWS)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAWS, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.CreateAwsIntegration(&secret_service.CreateAwsIntegrationParams{
			Body: &secretmodels.SecretServiceCreateAwsIntegrationBody{
				AccessKeys:                integration.accessKeys,
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
				Name:                      integration.Name.ValueString(),
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

func (r *resourceVaultSecretsIntegrationAWS) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAWS](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAWS)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAWS, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.UpdateAwsIntegration(&secret_service.UpdateAwsIntegrationParams{
			Body: &secretmodels.SecretServiceUpdateAwsIntegrationBody{
				AccessKeys:                integration.accessKeys,
				Capabilities:              integration.capabilities,
				FederatedWorkloadIdentity: integration.federatedWorkloadIdentity,
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

func (r *resourceVaultSecretsIntegrationAWS) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*IntegrationAWS](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		integration, ok := i.(*IntegrationAWS)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *IntegrationAWS, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecrets.DeleteAwsIntegration(
			secret_service.NewDeleteAwsIntegrationParamsWithContext(ctx).
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsIntegrationAWS) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Vault Secrets API does not return sensitive values like the secret access key, so they will be initialized to an empty value
	// It means the first plan/apply after a successful import will always show a diff for the secret access key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

var _ hvsResource = &IntegrationAWS{}

func (i *IntegrationAWS) projectID() types.String {
	return i.ProjectID
}

func (i *IntegrationAWS) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
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

	if !i.AccessKeys.IsNull() {
		ak := accessKeys{}
		diags = i.AccessKeys.As(ctx, &ak, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.accessKeys = &secretmodels.Secrets20231128AwsAccessKeysRequest{
			AccessKeyID:     ak.AccessKeyID.ValueString(),
			SecretAccessKey: ak.SecretAccessKey.ValueString(),
		}
	}

	if !i.FederatedWorkloadIdentity.IsNull() {
		fwi := awsFederatedWorkloadIdentity{}
		diags = i.FederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.federatedWorkloadIdentity = &secretmodels.Secrets20231128AwsFederatedWorkloadIdentityRequest{
			RoleArn:  fwi.RoleARN.ValueString(),
			Audience: fwi.Audience.ValueString(),
		}
	}

	return diag.Diagnostics{}
}

func (i *IntegrationAWS) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	integrationModel, ok := model.(*secretmodels.Secrets20231128AwsIntegration)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128AwsIntegration, got: %T", model))
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

	if integrationModel.AccessKeys != nil {
		// The secret key is not returned by the API, so we use an empty value (e.g. for imports) or the state value (e.g. for updates)
		secretAccessKey := ""
		if i.accessKeys != nil {
			secretAccessKey = i.accessKeys.SecretAccessKey
		}

		i.AccessKeys, diags = types.ObjectValue(i.AccessKeys.AttributeTypes(ctx), map[string]attr.Value{
			"access_key_id":     types.StringValue(integrationModel.AccessKeys.AccessKeyID),
			"secret_access_key": types.StringValue(secretAccessKey),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.FederatedWorkloadIdentity != nil {
		i.FederatedWorkloadIdentity, diags = types.ObjectValue(i.FederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"role_arn": types.StringValue(integrationModel.FederatedWorkloadIdentity.RoleArn),
			"audience": types.StringValue(integrationModel.FederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
