// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewRadarSecretManagerVaultDedicatedResource() resource.Resource {
	return &secretManagerResource{
		TypeName:          "_vault_radar_secret_manager_vault_dedicated",
		SecretManagerType: "hcp_vault_dedicated",
		ResourceSchema:    vaultDedicatedSchema,
		GetSecretManagerFromPlan: func(ctx context.Context, plan tfsdk.Plan) (secretManager, diag.Diagnostics) {
			var data vaultDedicatedModel
			diags := plan.Get(ctx, &data)
			return &data, diags
		},
		GetSecretManagerFromState: func(ctx context.Context, state tfsdk.State) (secretManager, diag.Diagnostics) {
			var data vaultDedicatedModel
			diags := state.Get(ctx, &data)
			return &data, diags
		}}
}

var vaultDedicatedSchema = schema.Schema{
	MarkdownDescription: "This terraform resource manages a HCP Vault Dedicated secret manager in Vault Radar.",
	Attributes: map[string]schema.Attribute{
		"vault_url": schema.StringAttribute{
			Description: `Specify the URL of the Vault instance without protocol. Example: 'acme-public-vault-abc.def.z1.hashicorp.cloud:8200'.`,
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-_.]+(:\d{2,})?$`),
					"url must contain only letters, numbers, hyphens, underscores, or periods, followed optionally by a colon and port number",
				),
			},
		},
		"access_read_write": schema.BoolAttribute{
			Description: `Indicates if the auth method has access has read and write access to the secrets engine paths. Defaults to false.`,
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"token": schema.SingleNestedAttribute{
			Description: `Configuration for token-based authentication. Only one authentication method may be configured.`,
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("kubernetes"),
					path.MatchRelative().AtParent().AtName("approle_push"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"token_env_var": schema.StringAttribute{
					Description: `Environment variable name containing the Vault token. Example: 'VAULT_TOKEN'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"token_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
			},
		},
		"kubernetes": schema.SingleNestedAttribute{
			Description: `Configuration for Kubernetes-based authentication. Only one authentication method may be configured.`,
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("token"),
					path.MatchRelative().AtParent().AtName("approle_push"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"mount_path": schema.StringAttribute{
					Description: `Mount path of the Kubernetes auth is enabled in Vault. Example 'kubernetes'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"role_name": schema.StringAttribute{
					Description: `Kubernetes authentication role configured in Vault.  Example 'vault-radar-role'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
		"approle_push": schema.SingleNestedAttribute{
			Description: `Configuration for AppRole Push-based authentication. Only one authentication method may be configured.`,
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("token"),
					path.MatchRelative().AtParent().AtName("kubernetes"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"mount_path": schema.StringAttribute{
					Description: `Mount path of the AppRole auth method in Vault. Example 'approle'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"role_id_env_var": schema.StringAttribute{
					Description: `Environment variable containing the AppRole role ID. Example: 'VAULT_APPROLE_ROLE_ID'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"role_id_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
				"secret_id_env_var": schema.StringAttribute{
					Description: `Environment variable containing the AppRole secret ID. Example: 'VAULT_APPROLE_SECRET_ID'.`,
					Required:    true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"secret_id_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
			},
		},
	},
}

type vaultDedicatedModel struct {
	abstractSecretManagerModel
	VaultURL          types.String       `tfsdk:"vault_url"`
	AccessReadWrite   types.Bool         `tfsdk:"access_read_write"`
	TokenConfig       *tokenConfig       `tfsdk:"token"`
	KubernetesConfig  *kubernetesConfig  `tfsdk:"kubernetes"`
	AppRolePushConfig *appRolePushConfig `tfsdk:"approle_push"`
}

type tokenConfig struct {
	TokenEnvVar types.String `tfsdk:"token_env_var"`
}

type kubernetesConfig struct {
	MountPath types.String `tfsdk:"mount_path"`
	RoleName  types.String `tfsdk:"role_name"`
}

type appRolePushConfig struct {
	MountPath      types.String `tfsdk:"mount_path"`
	RoleIDEnvVar   types.String `tfsdk:"role_id_env_var"`
	SecretIDEnvVar types.String `tfsdk:"secret_id_env_var"`
}

func (m *vaultDedicatedModel) GetConnectionURL() types.String { return m.VaultURL }

func (m *vaultDedicatedModel) SetFeatures(feature map[string]interface{}) {
	if feature == nil {
		m.AccessReadWrite = types.BoolValue(false)
		return
	}

	if val, ok := feature["copy_secrets"]; ok && val != nil {
		m.AccessReadWrite = types.BoolValue(true)
		return
	}

	m.AccessReadWrite = types.BoolValue(false)
}

func (m *vaultDedicatedModel) GetFeatures(omitEmptyValues bool) map[string]interface{} {
	if m.AccessReadWrite.IsNull() || m.AccessReadWrite.IsUnknown() || !m.AccessReadWrite.ValueBool() {
		if omitEmptyValues {
			// on create which uses a POST, we want to omit the copy_secrets field.
			return map[string]interface{}{}
		}
		// on update which uses a PATCH, we need to explicitly send copy_secrets field set to nil to clear it out.
		return map[string]interface{}{"copy_secrets": nil}
	}

	return map[string]interface{}{"copy_secrets": struct{}{}}
}

func (m *vaultDedicatedModel) GetAuthMethod() types.String {
	if m.TokenConfig != nil {
		return types.StringValue("token")
	}
	if m.KubernetesConfig != nil {
		return types.StringValue("kubernetes")
	}
	if m.AppRolePushConfig != nil {
		return types.StringValue("approle_push")
	}
	return types.StringNull()
}

func (m *vaultDedicatedModel) GetToken() types.String {
	if m.TokenConfig != nil {
		return basetypes.NewStringValue("env://" + m.TokenConfig.TokenEnvVar.ValueString())
	}

	if m.KubernetesConfig != nil {
		token_data, err := json.Marshal(map[string]interface{}{
			"type": "vault_kubernetes",
			"args": struct {
				ClusterType        string `json:"cluster_type"`
				ConnectionURL      string `json:"connection_url"`
				MountPath          string `json:"mount_path"`
				KubernetesAuthRole string `json:"kubernetes_auth_role"`
			}{
				ClusterType:        "hcp_vault_dedicated",
				ConnectionURL:      m.VaultURL.ValueString(),
				MountPath:          m.KubernetesConfig.MountPath.ValueString(),
				KubernetesAuthRole: m.KubernetesConfig.RoleName.ValueString(),
			},
		})

		if err != nil {
			tflog.Error(context.Background(), "error with auth_method kubernetes")
			return basetypes.NewStringNull()
		}

		return basetypes.NewStringValue(string(token_data))
	}

	if m.AppRolePushConfig != nil {
		token_data, err := json.Marshal(map[string]interface{}{
			"type": "vault_approle_push",
			"args": struct {
				ClusterType      string `json:"cluster_type"`
				ConnectionURL    string `json:"connection_url"`
				MountPath        string `json:"mount_path"`
				RoleIDLocation   string `json:"role_id_location"`
				SecretIDLocation string `json:"secret_id_location"`
			}{
				ClusterType:      "hcp_vault_dedicated",
				ConnectionURL:    m.VaultURL.ValueString(),
				MountPath:        m.AppRolePushConfig.MountPath.ValueString(),
				RoleIDLocation:   "env://" + m.AppRolePushConfig.RoleIDEnvVar.ValueString(),
				SecretIDLocation: "env://" + m.AppRolePushConfig.SecretIDEnvVar.ValueString(),
			},
		})

		if err != nil {
			tflog.Error(context.Background(), "error with auth_method approle_push")
			return basetypes.NewStringNull()
		}

		return basetypes.NewStringValue(string(token_data))
	}

	// No auth configuration provided
	tflog.Error(context.Background(), "no authentication configuration provided")
	return basetypes.NewStringNull()
}
