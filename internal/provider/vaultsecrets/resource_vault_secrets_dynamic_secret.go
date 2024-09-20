// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/maps"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

const (
	unsupportedProviderErrorFmt = "unsupported provider, expected one of %s, got '%s'"
	invalidSecretTypeErrorFmt   = "invalid secret type, expected *DynamicSecret, got: '%T', this is a bug on the provider"
)

var exactlyOneDynamicSecretTypeFieldsValidator = objectvalidator.ExactlyOneOf(
	path.Expressions{
		path.MatchRoot("aws_assume_role"),
		path.MatchRoot("gcp_impersonate_service_account"),
	}...,
)

// dynamicSecret encapsulates the HVS provider-specific logic so the Terraform resource can focus on the Terraform logic
type dynamicSecret interface {
	read(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error)
	create(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error)
	update(ctx context.Context, client secret_service.ClientService, secret *DynamicSecret) (any, error)
	// delete not necessary on the interface, all secrets use the same delete request
}

// dynamicSecretsImpl is a map of all the concrete dynamic secrets implementations by provider
// so the Terraform resource can look up the correct implementation based on the resource secret_provider field
var dynamicSecretsImpl = map[Provider]dynamicSecret{
	ProviderAWS: &awsDynamicSecret{},
	ProviderGCP: &gcpDynamicSecret{},
}

type DynamicSecret struct {
	// Shared input fields
	ProjectID       types.String `tfsdk:"project_id"`
	AppName         types.String `tfsdk:"app_name"`
	SecretProvider  types.String `tfsdk:"secret_provider"`
	Name            types.String `tfsdk:"name"`
	IntegrationName types.String `tfsdk:"integration_name"`
	DefaultTtl      types.String `tfsdk:"default_ttl"`

	// Provider specific mutually exclusive fields
	AWSAssumeRole                *awsAssumeRole                `tfsdk:"aws_assume_role"`
	GCPImpersonateServiceAccount *gcpImpersonateServiceAccount `tfsdk:"gcp_impersonate_service_account"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
}

type awsAssumeRole struct {
	IAMRoleARN types.String `tfsdk:"iam_role_arn"`
}

type gcpImpersonateServiceAccount struct {
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
}

var _ resource.Resource = &resourceVaultSecretsDynamicSecret{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsDynamicSecret{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsDynamicSecret{}

func NewVaultSecretsDynamicSecretResource() resource.Resource {
	return &resourceVaultSecretsDynamicSecret{}
}

type resourceVaultSecretsDynamicSecret struct {
	client *clients.Client
}

func (r *resourceVaultSecretsDynamicSecret) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_dynamic_secret"
}

func (r *resourceVaultSecretsDynamicSecret) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"default_ttl": schema.StringAttribute{
			Description: "TTL the generated credentials will be valid for.",
			Optional:    true,
		},
		"aws_assume_role": schema.SingleNestedAttribute{
			Description: "AWS configuration to generate dynamic credentials by assuming an IAM role. Required if `secret_provider` is `aws`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"iam_role_arn": schema.StringAttribute{
					Description: "AWS IAM role ARN to assume when generating credentials.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				exactlyOneDynamicSecretTypeFieldsValidator,
			},
		},
		"gcp_impersonate_service_account": schema.SingleNestedAttribute{
			Description: "GCP configuration to generate dynamic credentials by impersonating a service account. Required if `secret_provider` is `gcp`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"service_account_email": schema.StringAttribute{
					Description: "GCP service account email to impersonate.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				exactlyOneDynamicSecretTypeFieldsValidator,
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, managedSecretAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets dynamic secret resource manages a dynamic secret configuration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsDynamicSecret) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsDynamicSecret) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsDynamicSecret) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*DynamicSecret](ctx, r.client, &resp.State, req.State.Get, "reading", func(s hvsResource) (any, error) {
		secret, ok := s.(*DynamicSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		dynamicSecretImpl, ok := dynamicSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(dynamicSecretsImpl), secret.SecretProvider.ValueString())
		}
		return dynamicSecretImpl.read(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsDynamicSecret) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*DynamicSecret](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(s hvsResource) (any, error) {
		secret, ok := s.(*DynamicSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		dynamicSecretImpl, ok := dynamicSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(dynamicSecretsImpl), secret.SecretProvider.ValueString())
		}
		return dynamicSecretImpl.create(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsDynamicSecret) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*DynamicSecret](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(s hvsResource) (any, error) {
		secret, ok := s.(*DynamicSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		dynamicSecretImpl, ok := dynamicSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(dynamicSecretsImpl), secret.SecretProvider.ValueString())
		}
		return dynamicSecretImpl.update(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsDynamicSecret) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*DynamicSecret](ctx, r.client, &resp.State, req.State.Get, "deleting", func(s hvsResource) (any, error) {
		secret, ok := s.(*DynamicSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		_, err := r.client.VaultSecretsPreview.DeleteAppSecret(
			secret_service.NewDeleteAppSecretParamsWithContext(ctx).
				WithOrganizationID(secret.OrganizationID.ValueString()).
				WithProjectID(secret.ProjectID.ValueString()).
				WithAppName(secret.AppName.ValueString()).
				WithSecretName(secret.Name.ValueString()),
			nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

var _ hvsResource = &DynamicSecret{}

func (s *DynamicSecret) projectID() types.String {
	return s.ProjectID
}

func (s *DynamicSecret) initModel(_ context.Context, orgID, projID string) diag.Diagnostics {
	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)

	return diag.Diagnostics{}
}

func (s *DynamicSecret) fromModel(_ context.Context, orgID, projID string, _ any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)

	return diags
}
