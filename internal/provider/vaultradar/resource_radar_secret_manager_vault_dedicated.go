// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultradar

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			Description: `Specify the URL of the Vault instance without protocol. Example: acme-public-vault-abc.def.z1.hashicorp.cloud:8200`,
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
		"auth_method": schema.StringAttribute{
			Required:    true,
			Description: "The authentication method to use to connect to Vault. One of 'token', 'vault_kubernetes', or 'vault_approle_push'.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("token", "kubernetes", "approle_push"),
			},
		},
		"token": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"token_env_var": schema.StringAttribute{
					Description: `Environment variable name containing the Vault token. Example: 'VAULT_TOKEN'`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
						stringplanmodifier.UseStateForUnknown(),
					},
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"token_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
			},
		},
		"kubernetes": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"mount_path": schema.StringAttribute{
					Description: `Mount path of the Kubernetes auth is enabled in Vault. Example 'kubernetes'.`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					// TODO add validator not empty
				},
				"role_name": schema.StringAttribute{
					Description: `Kubernetes authentication role configured in Vault.`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					// TODO add validator not empty
				},
			},
		},
		"approle_push": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"mount_path": schema.StringAttribute{
					Description: `Mount path of the AppRole auth method in Vault. Example 'approle'.`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					// TODO add validator not empty
				},
				"role_id_env_var": schema.StringAttribute{
					Description: `Environment variable containing the AppRole role ID.`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"role_id_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
				"secret_id_env_var": schema.StringAttribute{
					Description: `Environment variable containing the AppRole secret ID.`,
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
							"secret_id_env_var must contain only letters, numbers, and underscores",
						),
					},
				},
			},
		},
		// TODO features, should this be a boolean or expose the actual features json?
	},
}

type vaultDedicatedModel struct {
	abstractSecretManagerModel
	VaultURL          types.String       `tfsdk:"vault_url"`
	AuthMethod        types.String       `tfsdk:"auth_method"`
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

func (m *vaultDedicatedModel) GetAuthMethod() types.String { return m.AuthMethod }

func (m *vaultDedicatedModel) GetToken() types.String {
	switch m.AuthMethod.ValueString() {
	case "token":
		if m.TokenConfig != nil {
			return basetypes.NewStringValue("env://" + m.TokenConfig.TokenEnvVar.ValueString())
		}
	case "kubernetes":
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

		// else no kubernetes config
		tflog.Error(context.Background(), "error no data for auth_method kubernetes")
	case "approle_push":
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

		// else no approle_push config
		tflog.Error(context.Background(), "error no data for auth_method approle_push")
	}

	// else no auth_method match or no config
	tflog.Error(context.Background(), "error no data for auth_method")
	return basetypes.NewStringNull()
}
