// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

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
	"golang.org/x/exp/maps"

	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
)

var exactlyOneIntegrationTypeFieldsValidator = objectvalidator.ExactlyOneOf(
	path.Expressions{
		path.MatchRoot("aws_access_keys"),
		path.MatchRoot("aws_federated_workload_identity"),
		path.MatchRoot("azure_client_secret"),
		path.MatchRoot("azure_federated_workload_identity"),
		path.MatchRoot("confluent_static_credentials"),
		path.MatchRoot("gcp_service_account_key"),
		path.MatchRoot("gcp_federated_workload_identity"),
		path.MatchRoot("mongodb_atlas_static_credentials"),
		path.MatchRoot("twilio_static_credentials"),
		path.MatchRoot("mssql_static_credentials"),
	}...,
)

type mssqlStaticCredentials struct {
	ConnectionString types.String `tfsdk:"connection_string"`
}

type Integration struct {
	// Input fields
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Capabilities types.Set    `tfsdk:"capabilities"`
	Provider     types.String `tfsdk:"provider_type"`

	// Provider specific mutually exclusive fields
	AwsAccessKeys                    types.Object `tfsdk:"aws_access_keys"`
	AwsFederatedWorkloadIdentity     types.Object `tfsdk:"aws_federated_workload_identity"`
	AzureClientSecret                types.Object `tfsdk:"azure_client_secret"`
	AzureFederatedWorkloadIdentity   types.Object `tfsdk:"azure_federated_workload_identity"`
	ConfluentStaticCredentialDetails types.Object `tfsdk:"confluent_static_credentials"`
	GcpServiceAccountKey             types.Object `tfsdk:"gcp_service_account_key"`
	GcpFederatedWorkloadIdentity     types.Object `tfsdk:"gcp_federated_workload_identity"`
	MongoDBAtlasStaticCredentials    types.Object `tfsdk:"mongodb_atlas_static_credentials"`
	TwilioStaticCredentials          types.Object `tfsdk:"twilio_static_credentials"`
	MssqlStaticCredentials           types.Object `tfsdk:"mssql_static_credentials"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceName   types.String `tfsdk:"resource_name"`

	// Inner API-compatible models derived from the Terraform fields
	capabilities                   []*secretmodels.Secrets20231128Capability                          `tfsdk:"-"`
	awsAccessKeys                  *secretmodels.Secrets20231128AwsAccessKeysRequest                  `tfsdk:"-"`
	awsFederatedWorkloadIdentity   *secretmodels.Secrets20231128AwsFederatedWorkloadIdentityRequest   `tfsdk:"-"`
	azureClientSecret              *secretmodels.Secrets20231128AzureClientSecretRequest              `tfsdk:"-"`
	azureFederatedWorkloadIdentity *secretmodels.Secrets20231128AzureFederatedWorkloadIdentityRequest `tfsdk:"-"`
	confluentStaticCredentials     *secretmodels.Secrets20231128ConfluentStaticCredentialsRequest     `tfsdk:"-"`
	gcpServiceAccountKey           *secretmodels.Secrets20231128GcpServiceAccountKeyRequest           `tfsdk:"-"`
	gcpFederatedWorkloadIdentity   *secretmodels.Secrets20231128GcpFederatedWorkloadIdentityRequest   `tfsdk:"-"`
	mongoDBAtlasStaticCredentials  *secretmodels.Secrets20231128MongoDBAtlasStaticCredentialsRequest  `tfsdk:"-"`
	twilioStaticCredentials        *secretmodels.Secrets20231128TwilioStaticCredentialsRequest        `tfsdk:"-"`
	mssqlStaticCredentials         *secretmodels.Secrets20231128MssqlStaticCredentialsRequest         `tfsdk:"-"`
}

var _ resource.Resource = &resourceVaultSecretsIntegration{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsIntegration{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsIntegration{}
var _ resource.ResourceWithImportState = &resourceVaultSecretsIntegration{}

func NewVaultSecretsIntegrationResource() resource.Resource {
	return &resourceVaultSecretsIntegration{}
}

type resourceVaultSecretsIntegration struct {
	client *clients.Client
}

func (r *resourceVaultSecretsIntegration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_integration"
}

func (r *resourceVaultSecretsIntegration) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"provider_type": schema.StringAttribute{
			Description: "The provider or 3rd party platform the integration is for.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(ProviderStrings()...),
			},
		},
		"aws_access_keys": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"aws_federated_workload_identity": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"azure_client_secret": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"azure_federated_workload_identity": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"confluent_static_credentials": schema.SingleNestedAttribute{
			Description: "Confluent API key used to authenticate for cloud apis.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"cloud_api_key_id": schema.StringAttribute{
					Description: "Public key used alongside the private key to authenticate for cloud apis.",
					Required:    true,
				},
				"cloud_api_secret": schema.StringAttribute{
					Description: "Private key used alongside the public key to authenticate for cloud apis.",
					Required:    true,
					Sensitive:   true,
				},
			},
			Validators: []validator.Object{
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"gcp_service_account_key": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"gcp_federated_workload_identity": schema.SingleNestedAttribute{
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
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"mongodb_atlas_static_credentials": schema.SingleNestedAttribute{
			Description: "MongoDB Atlas API key used to authenticate against the target project.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"api_public_key": schema.StringAttribute{
					Description: "Public key used alongside the private key to authenticate against the target project.",
					Required:    true,
				},
				"api_private_key": schema.StringAttribute{
					Description: "Private key used alongside the public key to authenticate against the target project.",
					Required:    true,
					Sensitive:   true,
				},
			},
			Validators: []validator.Object{
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"twilio_static_credentials": schema.SingleNestedAttribute{
			Description: "Twilio API key parts used to authenticate against the target Twilio account.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"account_sid": schema.StringAttribute{
					Description: "Account SID for the target Twilio account.",
					Required:    true,
				},
				"api_key_sid": schema.StringAttribute{
					Description: "Api key SID to authenticate against the target Twilio account.",
					Required:    true,
				},
				"api_key_secret": schema.StringAttribute{
					Description: "Api key secret used with the api key SID to authenticate against the target Twilio account.",
					Required:    true,
					Sensitive:   true,
				},
			},
			Validators: []validator.Object{
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
		"mssql_static_credentials": schema.SingleNestedAttribute{
			Description: "MsSQL API key parts used to authenticate against the target MsSQL account.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"connection_string": schema.StringAttribute{
					Description: "Connection string for the target MsSQL account.",
					Required:    true,
				},
			},
			Validators: []validator.Object{
				exactlyOneIntegrationTypeFieldsValidator,
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, sharedIntegrationAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets integration resource manages an integration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsIntegration) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsIntegration) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*Integration](ctx, r.client, &resp.State, req.State.Get, "reading", func(i hvsResource) (any, error) {
		integration, ok := i.(*Integration)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Integration, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.GetIntegration(
			secret_service.NewGetIntegrationParamsWithContext(ctx).
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

func (r *resourceVaultSecretsIntegration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*Integration](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(i hvsResource) (any, error) {
		integration, ok := i.(*Integration)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Integration, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.CreateIntegration(&secret_service.CreateIntegrationParams{
			Body: &secretmodels.SecretServiceCreateIntegrationBody{
				Name:                           integration.Name.ValueString(),
				Provider:                       integration.Provider.ValueString(),
				Capabilities:                   integration.capabilities,
				AwsAccessKeys:                  integration.awsAccessKeys,
				AwsFederatedWorkloadIdentity:   integration.awsFederatedWorkloadIdentity,
				AzureClientSecret:              integration.azureClientSecret,
				AzureFederatedWorkloadIdentity: integration.azureFederatedWorkloadIdentity,
				ConfluentStaticCredentials:     integration.confluentStaticCredentials,
				GcpServiceAccountKey:           integration.gcpServiceAccountKey,
				GcpFederatedWorkloadIdentity:   integration.gcpFederatedWorkloadIdentity,
				MongoDbAtlasStaticCredentials:  integration.mongoDBAtlasStaticCredentials,
				TwilioStaticCredentials:        integration.twilioStaticCredentials,
				MssqlStaticCredentials:         integration.mssqlStaticCredentials,
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

func (r *resourceVaultSecretsIntegration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*Integration](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(i hvsResource) (any, error) {
		integration, ok := i.(*Integration)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Integration, got: %T, this is a bug on the provider", i)
		}

		response, err := r.client.VaultSecrets.UpdateIntegration(&secret_service.UpdateIntegrationParams{
			Body: &secretmodels.SecretServiceUpdateIntegrationBody{
				Provider:                       integration.Provider.ValueString(),
				Capabilities:                   integration.capabilities,
				AwsAccessKeys:                  integration.awsAccessKeys,
				AwsFederatedWorkloadIdentity:   integration.awsFederatedWorkloadIdentity,
				AzureClientSecret:              integration.azureClientSecret,
				AzureFederatedWorkloadIdentity: integration.azureFederatedWorkloadIdentity,
				ConfluentStaticCredentials:     integration.confluentStaticCredentials,
				GcpServiceAccountKey:           integration.gcpServiceAccountKey,
				GcpFederatedWorkloadIdentity:   integration.gcpFederatedWorkloadIdentity,
				MongoDbAtlasStaticCredentials:  integration.mongoDBAtlasStaticCredentials,
				TwilioStaticCredentials:        integration.twilioStaticCredentials,
				MssqlStaticCredentials:         integration.mssqlStaticCredentials,
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

func (r *resourceVaultSecretsIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*Integration](ctx, r.client, &resp.State, req.State.Get, "deleting", func(i hvsResource) (any, error) {
		integration, ok := i.(*Integration)
		if !ok {
			return nil, fmt.Errorf("invalid integration type, expected *Integration, got: %T, this is a bug on the provider", i)
		}

		_, err := r.client.VaultSecrets.DeleteIntegration(
			secret_service.NewDeleteIntegrationParamsWithContext(ctx).
				WithOrganizationID(integration.OrganizationID.ValueString()).
				WithProjectID(integration.ProjectID.ValueString()).
				WithName(integration.Name.ValueString()), nil)
		if err != nil && !clients.IsResponseCodeNotFound(err) {
			return nil, err
		}
		return nil, nil
	})...)
}

func (r *resourceVaultSecretsIntegration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The Vault Secrets API does not return sensitive values like the secret access key, so they will be initialized to an empty value
	// It means the first plan/apply after a successful import will always show a diff for the secret access key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), r.client.Config.OrganizationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), r.client.Config.ProjectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

var _ hvsResource = &Integration{}

func (i *Integration) projectID() types.String {
	return i.ProjectID
}

func (i *Integration) initModel(ctx context.Context, orgID, projID string) diag.Diagnostics {
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

	if !i.AwsAccessKeys.IsNull() {
		ak := accessKeys{}
		diags = i.AwsAccessKeys.As(ctx, &ak, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.awsAccessKeys = &secretmodels.Secrets20231128AwsAccessKeysRequest{
			AccessKeyID:     ak.AccessKeyID.ValueString(),
			SecretAccessKey: ak.SecretAccessKey.ValueString(),
		}
	}

	if !i.AwsFederatedWorkloadIdentity.IsNull() {
		fwi := awsFederatedWorkloadIdentity{}
		diags = i.AwsFederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.awsFederatedWorkloadIdentity = &secretmodels.Secrets20231128AwsFederatedWorkloadIdentityRequest{
			RoleArn:  fwi.RoleARN.ValueString(),
			Audience: fwi.Audience.ValueString(),
		}
	}

	if !i.AzureClientSecret.IsNull() {
		cs := clientSecret{}
		diags = i.AzureClientSecret.As(ctx, &cs, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.azureClientSecret = &secretmodels.Secrets20231128AzureClientSecretRequest{
			TenantID:     cs.TenantID.ValueString(),
			ClientID:     cs.ClientID.ValueString(),
			ClientSecret: cs.ClientSecret.ValueString(),
		}
	}

	if !i.AzureFederatedWorkloadIdentity.IsNull() {
		fwi := azureFederatedWorkloadIdentity{}
		diags = i.AzureFederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.azureFederatedWorkloadIdentity = &secretmodels.Secrets20231128AzureFederatedWorkloadIdentityRequest{
			Audience: fwi.Audience.ValueString(),
			TenantID: fwi.TenantID.ValueString(),
			ClientID: fwi.ClientID.ValueString(),
		}
	}

	if !i.ConfluentStaticCredentialDetails.IsNull() {
		scd := confluentStaticCredentialDetails{}
		diags = i.ConfluentStaticCredentialDetails.As(ctx, &scd, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.confluentStaticCredentials = &secretmodels.Secrets20231128ConfluentStaticCredentialsRequest{
			CloudAPIKeyID:  scd.CloudAPIKeyID.ValueString(),
			CloudAPISecret: scd.CloudAPISecret.ValueString(),
		}
	}

	if !i.GcpServiceAccountKey.IsNull() {
		sa := serviceAccountKey{}
		diags = i.GcpServiceAccountKey.As(ctx, &sa, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.gcpServiceAccountKey = &secretmodels.Secrets20231128GcpServiceAccountKeyRequest{
			Credentials: sa.Credentials.ValueString(),
		}
	}

	if !i.GcpFederatedWorkloadIdentity.IsNull() {
		fwi := gcpFederatedWorkloadIdentity{}
		diags = i.GcpFederatedWorkloadIdentity.As(ctx, &fwi, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.gcpFederatedWorkloadIdentity = &secretmodels.Secrets20231128GcpFederatedWorkloadIdentityRequest{
			Audience:            fwi.Audience.ValueString(),
			ServiceAccountEmail: fwi.ServiceAccountEmail.ValueString(),
		}
	}

	if !i.MongoDBAtlasStaticCredentials.IsNull() {
		scd := mongoDBAtlasStaticCredentialDetails{}
		diags = i.MongoDBAtlasStaticCredentials.As(ctx, &scd, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.mongoDBAtlasStaticCredentials = &secretmodels.Secrets20231128MongoDBAtlasStaticCredentialsRequest{
			APIPublicKey:  scd.APIPublicKey.ValueString(),
			APIPrivateKey: scd.APIPrivateKey.ValueString(),
		}
	}

	if !i.TwilioStaticCredentials.IsNull() {
		scd := staticCredentialDetails{}
		diags = i.TwilioStaticCredentials.As(ctx, &scd, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.twilioStaticCredentials = &secretmodels.Secrets20231128TwilioStaticCredentialsRequest{
			AccountSid:   scd.AccountSID.ValueString(),
			APIKeySecret: scd.APIKeySecret.ValueString(),
			APIKeySid:    scd.APIKeySID.ValueString(),
		}
	}

	if !i.MssqlStaticCredentials.IsNull() {
		scd := mssqlStaticCredentials{}
		diags = i.MssqlStaticCredentials.As(ctx, &scd, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		i.mssqlStaticCredentials = &secretmodels.Secrets20231128MssqlStaticCredentialsRequest{
			ConnectionString: scd.ConnectionString.ValueString(),
		}
	}

	return diag.Diagnostics{}
}

func (i *Integration) fromModel(ctx context.Context, orgID, projID string, model any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	integrationModel, ok := model.(*secretmodels.Secrets20231128Integration)
	if !ok {
		diags.AddError("Invalid model type, this is a bug on the provider.", fmt.Sprintf("Expected *secretmodels.Secrets20231128Integration, got: %T", model))
		return diags
	}

	i.OrganizationID = types.StringValue(orgID)
	i.ProjectID = types.StringValue(projID)
	i.ResourceID = types.StringValue(integrationModel.ResourceID)
	i.ResourceName = types.StringValue(integrationModel.ResourceName)
	i.Name = types.StringValue(integrationModel.Name)
	i.Provider = types.StringValue(integrationModel.Provider)

	var values []attr.Value
	for _, c := range integrationModel.Capabilities {
		values = append(values, types.StringValue(string(*c)))
	}
	i.Capabilities, diags = types.SetValue(types.StringType, values)
	if diags.HasError() {
		return diags
	}

	if integrationModel.AwsAccessKeys != nil {
		secretAccessKey := ""
		if i.awsAccessKeys != nil {
			secretAccessKey = i.awsAccessKeys.SecretAccessKey
		}

		i.AwsAccessKeys, diags = types.ObjectValue(i.AwsAccessKeys.AttributeTypes(ctx), map[string]attr.Value{
			"access_key_id":     types.StringValue(integrationModel.AwsAccessKeys.AccessKeyID),
			"secret_access_key": types.StringValue(secretAccessKey),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.AwsFederatedWorkloadIdentity != nil {
		i.AwsFederatedWorkloadIdentity, diags = types.ObjectValue(i.AwsFederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"role_arn": types.StringValue(integrationModel.AwsFederatedWorkloadIdentity.RoleArn),
			"audience": types.StringValue(integrationModel.AwsFederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.AzureClientSecret != nil {
		clientSecret := ""
		if i.azureClientSecret != nil {
			clientSecret = i.azureClientSecret.ClientSecret
		}
		i.AzureClientSecret, diags = types.ObjectValue(i.AzureClientSecret.AttributeTypes(ctx), map[string]attr.Value{
			"tenant_id":     types.StringValue(integrationModel.AzureClientSecret.TenantID),
			"client_id":     types.StringValue(integrationModel.AzureClientSecret.ClientID),
			"client_secret": types.StringValue(clientSecret),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.AzureFederatedWorkloadIdentity != nil {
		i.AzureFederatedWorkloadIdentity, diags = types.ObjectValue(i.AzureFederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"tenant_id": types.StringValue(integrationModel.AzureFederatedWorkloadIdentity.TenantID),
			"client_id": types.StringValue(integrationModel.AzureFederatedWorkloadIdentity.ClientID),
			"audience":  types.StringValue(integrationModel.AzureFederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.ConfluentStaticCredentials != nil {
		cloudAPISecret := ""
		if i.confluentStaticCredentials != nil {
			cloudAPISecret = i.confluentStaticCredentials.CloudAPISecret
		}

		i.ConfluentStaticCredentialDetails, diags = types.ObjectValue(i.ConfluentStaticCredentialDetails.AttributeTypes(ctx), map[string]attr.Value{
			"cloud_api_key_id": types.StringValue(integrationModel.ConfluentStaticCredentials.CloudAPIKeyID),
			"cloud_api_secret": types.StringValue(cloudAPISecret),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.GcpServiceAccountKey != nil {
		credentials := ""
		if i.gcpServiceAccountKey != nil {
			credentials = i.gcpServiceAccountKey.Credentials
		}

		i.GcpServiceAccountKey, diags = types.ObjectValue(i.GcpServiceAccountKey.AttributeTypes(ctx), map[string]attr.Value{
			"credentials":  types.StringValue(credentials),
			"project_id":   types.StringValue(integrationModel.GcpServiceAccountKey.ProjectID),
			"client_email": types.StringValue(integrationModel.GcpServiceAccountKey.ClientEmail),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.GcpFederatedWorkloadIdentity != nil {
		i.GcpFederatedWorkloadIdentity, diags = types.ObjectValue(i.GcpFederatedWorkloadIdentity.AttributeTypes(ctx), map[string]attr.Value{
			"service_account_email": types.StringValue(integrationModel.GcpFederatedWorkloadIdentity.ServiceAccountEmail),
			"audience":              types.StringValue(integrationModel.GcpFederatedWorkloadIdentity.Audience),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.MongoDbAtlasStaticCredentials != nil {
		apiPrivateKey := ""
		if i.mongoDBAtlasStaticCredentials != nil {
			apiPrivateKey = i.mongoDBAtlasStaticCredentials.APIPrivateKey
		}

		i.MongoDBAtlasStaticCredentials, diags = types.ObjectValue(i.MongoDBAtlasStaticCredentials.AttributeTypes(ctx), map[string]attr.Value{
			"api_public_key":  types.StringValue(integrationModel.MongoDbAtlasStaticCredentials.APIPublicKey),
			"api_private_key": types.StringValue(apiPrivateKey),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.TwilioStaticCredentials != nil {
		apiKeySecret := ""
		if i.twilioStaticCredentials != nil {
			apiKeySecret = i.twilioStaticCredentials.APIKeySecret
		}

		i.TwilioStaticCredentials, diags = types.ObjectValue(i.TwilioStaticCredentials.AttributeTypes(ctx), map[string]attr.Value{
			"account_sid":    types.StringValue(integrationModel.TwilioStaticCredentials.AccountSid),
			"api_key_sid":    types.StringValue(integrationModel.TwilioStaticCredentials.APIKeySid),
			"api_key_secret": types.StringValue(apiKeySecret),
		})
		if diags.HasError() {
			return diags
		}
	}

	if integrationModel.MssqlStaticCredentials != nil {
		i.MssqlStaticCredentials, diags = types.ObjectValue(i.MssqlStaticCredentials.AttributeTypes(ctx), map[string]attr.Value{
			"connection_string": types.StringValue(integrationModel.MssqlStaticCredentials.ConnectionString),
		})
		if diags.HasError() {
			return diags
		}
	}

	return diags
}
