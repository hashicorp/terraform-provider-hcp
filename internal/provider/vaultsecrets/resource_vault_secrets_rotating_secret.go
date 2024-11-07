// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	secretmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/modifiers"
	"golang.org/x/exp/maps"
)

var exactlyOneRotatingSecretTypeFieldsValidator = objectvalidator.ExactlyOneOf(
	path.Expressions{
		path.MatchRoot("aws_access_keys"),
		path.MatchRoot("gcp_service_account_key"),
		path.MatchRoot("mongodb_atlas_user"),
		path.MatchRoot("twilio_api_key"),
		path.MatchRoot("confluent_service_account"),
		path.MatchRoot("postgres_usernames"),
	}...,
)

// rotatingSecret encapsulates the HVS provider-specific logic so the Terraform resource can focus on the Terraform logic
type rotatingSecret interface {
	read(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error)
	create(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error)
	update(ctx context.Context, client secret_service.ClientService, secret *RotatingSecret) (any, error)
	// delete not necessary on the interface, all secrets use the same delete request
}

// rotatingSecretsImpl is a map of all the concrete rotating secrets implementations by provider
// so the Terraform resource can look up the correct implementation based on the resource secret_provider field
var rotatingSecretsImpl = map[Provider]rotatingSecret{
	ProviderAWS:          &awsRotatingSecret{},
	ProviderGCP:          &gcpRotatingSecret{},
	ProviderMongoDBAtlas: &mongoDBAtlasRotatingSecret{},
	ProviderTwilio:       &twilioRotatingSecret{},
	ProviderConfluent:    &confluentRotatingSecret{},
	ProviderPostgres:     &postgresRotatingSecret{},
}

type RotatingSecret struct {
	// Shared input fields
	ProjectID          types.String `tfsdk:"project_id"`
	AppName            types.String `tfsdk:"app_name"`
	SecretProvider     types.String `tfsdk:"secret_provider"`
	Name               types.String `tfsdk:"name"`
	IntegrationName    types.String `tfsdk:"integration_name"`
	RotationPolicyName types.String `tfsdk:"rotation_policy_name"`

	// Provider specific mutually exclusive fields
	AWSAccessKeys           *awsAccessKeys           `tfsdk:"aws_access_keys"`
	GCPServiceAccountKey    *gcpServiceAccountKey    `tfsdk:"gcp_service_account_key"`
	MongoDBAtlasUser        *mongoDBAtlasUser        `tfsdk:"mongodb_atlas_user"`
	TwilioAPIKey            *twilioAPIKey            `tfsdk:"twilio_api_key"`
	ConfluentServiceAccount *confluentServiceAccount `tfsdk:"confluent_service_account"`
	PostgresUsernames       *postgresUsernames       `tfsdk:"postgres_usernames"`

	// Computed fields
	OrganizationID types.String `tfsdk:"organization_id"`

	// Inner API-compatible models derived from the Terraform fields
	mongoDBRoles []*secretmodels.Secrets20231128MongoDBRole `tfsdk:"-"`
}

type awsAccessKeys struct {
	IAMUsername types.String `tfsdk:"iam_username"`
}

type gcpServiceAccountKey struct {
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
}

type mongoDBAtlasUser struct {
	ProjectID    types.String   `tfsdk:"project_id"`
	DatabaseName types.String   `tfsdk:"database_name"`
	Roles        []types.String `tfsdk:"roles"`
}

type confluentServiceAccount struct {
	ServiceAccountID types.String `tfsdk:"service_account_id"`
}

type twilioAPIKey struct{}

type postgresUsernames struct {
	Usernames []types.String `tfsdk:"usernames"`
}

var _ resource.Resource = &resourceVaultSecretsRotatingSecret{}
var _ resource.ResourceWithConfigure = &resourceVaultSecretsRotatingSecret{}
var _ resource.ResourceWithModifyPlan = &resourceVaultSecretsRotatingSecret{}

func NewVaultSecretsRotatingSecretResource() resource.Resource {
	return &resourceVaultSecretsRotatingSecret{}
}

type resourceVaultSecretsRotatingSecret struct {
	client *clients.Client
}

func (r *resourceVaultSecretsRotatingSecret) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_rotating_secret"
}

func (r *resourceVaultSecretsRotatingSecret) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"rotation_policy_name": schema.StringAttribute{
			Description: "Name of the rotation policy that governs the rotation of the secret.",
			Required:    true,
		},
		"aws_access_keys": schema.SingleNestedAttribute{
			Description: "AWS configuration to manage the access key rotation for the given IAM user. Required if `secret_provider` is `aws`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"iam_username": schema.StringAttribute{
					Description: "AWS IAM username to rotate the access keys for.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
		"gcp_service_account_key": schema.SingleNestedAttribute{
			Description: "GCP configuration to manage the service account key rotation for the given service account. Required if `secret_provider` is `gcp`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"service_account_email": schema.StringAttribute{
					Description: "GCP service account email to impersonate.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
		"mongodb_atlas_user": schema.SingleNestedAttribute{
			Description: "MongoDB Atlas configuration to manage the user password rotation on the given database. Required if `secret_provider` is `mongodb_atlas`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"project_id": schema.StringAttribute{
					Description: "MongoDB Atlas project ID to rotate the username and password for.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"database_name": schema.StringAttribute{
					Description: "MongoDB Atlas database or cluster name to rotate the username and password for.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"roles": schema.ListAttribute{
					Description: "MongoDB Atlas roles to assign to the rotating user.",
					Required:    true,
					ElementType: types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
		"twilio_api_key": schema.SingleNestedAttribute{
			Description: "Twilio configuration to manage the api key rotation on the given account. Required if `secret_provider` is `twilio`.",
			Optional:    true,
			Attributes:  map[string]schema.Attribute{
				// Twilio does not have rotating-secret-specific fields for the moment, this block is to preserve future backwards compatibility
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
		"confluent_service_account": schema.SingleNestedAttribute{
			Description: "Confluent configuration to manage the cloud api key rotation for the given service account. Required if `secret_provider` is `confluent`.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"service_account_id": schema.StringAttribute{
					Description: "Confluent service account to rotate the cloud api key for.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
		"postgres_usernames": schema.SingleNestedAttribute{
			Description: "",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"usernames": schema.ListAttribute{
					Description: "Postgres usernames to rotate passwords for.",
					Required:    true,
					ElementType: types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				exactlyOneRotatingSecretTypeFieldsValidator,
			},
		},
	}

	maps.Copy(attributes, locationAttributes)
	maps.Copy(attributes, managedSecretAttributes)

	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets rotating secret resource manages a rotating secret configuration.",
		Attributes:          attributes,
	}
}

func (r *resourceVaultSecretsRotatingSecret) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceVaultSecretsRotatingSecret) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifiers.ModifyPlanForDefaultProjectChange(ctx, r.client.Config.ProjectID, req.State, req.Config, req.Plan, resp)
}

func (r *resourceVaultSecretsRotatingSecret) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(decorateOperation[*RotatingSecret](ctx, r.client, &resp.State, req.State.Get, "reading", func(s hvsResource) (any, error) {
		secret, ok := s.(*RotatingSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		rotatingSecretImpl, ok := rotatingSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(rotatingSecretsImpl), secret.SecretProvider.ValueString())
		}
		return rotatingSecretImpl.read(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsRotatingSecret) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(decorateOperation[*RotatingSecret](ctx, r.client, &resp.State, req.Plan.Get, "creating", func(s hvsResource) (any, error) {
		secret, ok := s.(*RotatingSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		rotatingSecretImpl, ok := rotatingSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(rotatingSecretsImpl), secret.SecretProvider.ValueString())
		}
		return rotatingSecretImpl.create(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsRotatingSecret) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(decorateOperation[*RotatingSecret](ctx, r.client, &resp.State, req.Plan.Get, "updating", func(s hvsResource) (any, error) {
		secret, ok := s.(*RotatingSecret)
		if !ok {
			return nil, fmt.Errorf(invalidSecretTypeErrorFmt, s)
		}

		rotatingSecretImpl, ok := rotatingSecretsImpl[Provider(secret.SecretProvider.ValueString())]
		if !ok {
			return nil, fmt.Errorf(unsupportedProviderErrorFmt, maps.Keys(rotatingSecretsImpl), secret.SecretProvider.ValueString())
		}
		return rotatingSecretImpl.update(ctx, r.client.VaultSecretsPreview, secret)
	})...)
}

func (r *resourceVaultSecretsRotatingSecret) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(decorateOperation[*RotatingSecret](ctx, r.client, &resp.State, req.State.Get, "deleting", func(s hvsResource) (any, error) {
		secret, ok := s.(*RotatingSecret)
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

var _ hvsResource = &RotatingSecret{}

func (s *RotatingSecret) projectID() types.String {
	return s.ProjectID
}

func (s *RotatingSecret) initModel(_ context.Context, orgID, projID string) diag.Diagnostics {
	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)

	if s.MongoDBAtlasUser != nil {
		for _, r := range s.MongoDBAtlasUser.Roles {
			role := secretmodels.Secrets20231128MongoDBRole{
				DatabaseName: s.MongoDBAtlasUser.DatabaseName.ValueString(),
				RoleName:     r.ValueString(),
			}
			s.mongoDBRoles = append(s.mongoDBRoles, &role)
		}
	}

	return diag.Diagnostics{}
}

func (s *RotatingSecret) fromModel(_ context.Context, orgID, projID string, _ any) diag.Diagnostics {
	diags := diag.Diagnostics{}

	s.OrganizationID = types.StringValue(orgID)
	s.ProjectID = types.StringValue(projID)

	return diags
}
