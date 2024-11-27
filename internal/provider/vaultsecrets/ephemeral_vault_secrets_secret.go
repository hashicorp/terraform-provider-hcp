package vaultsecrets

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type ephemeralVaultSecretsSecretModel struct {
	ID          types.String `tfsdk:"id"`
	AppName     types.String `tfsdk:"app_name"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
	SecretName  types.String `tfsdk:"secret_name"`
	SecretValue types.String `tfsdk:"secret_value"`
}

var _ ephemeral.EphemeralResource = &ephemeralVaultSecretsSecret{}

type ephemeralVaultSecretsSecret struct {
	client *clients.Client
}

func (e *ephemeralVaultSecretsSecret) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_secret"
}

func (e *ephemeralVaultSecretsSecret) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema.Description = "This ephemeral value provides a value retrieved from a singular secret at its latest version."
	resp.Schema.MarkdownDescription = "This ephemeral value provides a value retrieved from a singular secret at its latest version."

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
			"app_name": schema.StringAttribute{
				Description: "The name of the Vault Secrets application.",
				Required:    true,
			},
			"secret_name": schema.StringAttribute{
				Description: "The name of the Vault Secrets secret.",
				Required:    true,
			},
			"secret_value": schema.StringAttribute{
				Description: "The secret value corresponding to the secret name input.",
				Computed:    true,
				Sensitive:   true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Vault Secrets app is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Vault Secrets app is located.",
				Computed:    true,
			},
		},
	}
}

func (e *ephemeralVaultSecretsSecret) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source client type",
			fmt.Sprintf("Expected *clients.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	e.client = pd
}

func (e *ephemeralVaultSecretsSecret) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ephemeralVaultSecretsSecretModel
	payload := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(payload...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := e.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: e.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	appSecret, err := clients.OpenVaultSecretsAppSecret(ctx, e.client, loc, data.AppName.ValueString(), data.SecretName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Error reading secret")
		return
	}

	secretValue := types.StringValue(appSecret.StaticVersion.Value)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &secretValue)...)
}

func (e *ephemeralVaultSecretsSecret) ConfigValidators(_ context.Context) []ephemeral.ConfigValidator {
	return nil
}
