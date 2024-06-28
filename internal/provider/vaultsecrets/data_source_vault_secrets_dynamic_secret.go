// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceVaultSecretsDynamicSecret struct {
	client *clients.Client
}

type DataSourceVaultSecretsDynamicSecretModel struct {
	// Config fields
	AppName    types.String `tfsdk:"app_name"`
	SecretName types.String `tfsdk:"secret_name"`

	// Data fields
	OrgID          types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
	SecretProvider types.String `tfsdk:"secret_provider"`
	SecretValues   types.Map    `tfsdk:"secret_values"`
}

func NewVaultSecretsDynamicSecretDataSource() datasource.DataSource {
	return &DataSourceVaultSecretsDynamicSecret{}
}

func (d *DataSourceVaultSecretsDynamicSecret) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_dynamic_secret"
}

func (d *DataSourceVaultSecretsDynamicSecret) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source generates a new dynamic secret instance.",
		Attributes: map[string]schema.Attribute{
			"app_name": schema.StringAttribute{
				Description: "The name of the Vault Secrets application.",
				Required:    true,
			},
			"secret_name": schema.StringAttribute{
				Description: "The name of the Vault Secrets secret.",
				Required:    true,
			},
			"secret_values": schema.MapAttribute{
				Description: "The secret values corresponding to the secret name input.",
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"secret_provider": schema.StringAttribute{
				Description: "The name of the provider this dynamic secret is for",
				Computed:    true,
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

func (d *DataSourceVaultSecretsDynamicSecret) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}

func (d *DataSourceVaultSecretsDynamicSecret) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceVaultSecretsDynamicSecretModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	client := d.client
	if client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HCP Client",
			"Expected configured HCP client. Please report this issue to the provider developers.",
		)
		return
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: client.Config.OrganizationID,
		ProjectID:      client.Config.ProjectID,
	}

	openSecret, err := clients.OpenVaultSecretsAppSecret(ctx, client, loc, data.AppName.ValueString(), data.SecretName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "Unable to open secret")
		return
	}

	var secretValues map[string]string
	switch {
	case openSecret.DynamicInstance != nil:
		secretValues = openSecret.DynamicInstance.Values
	default:
		resp.Diagnostics.AddError(
			"Unsupported HCP Secret type",
			fmt.Sprintf("HCP Secrets secret type %q is not currently supported by terraform-provider-hcp", openSecret.Type),
		)
		return
	}

	secretsOutput, diag := types.MapValueFrom(ctx, types.StringType, secretValues)
	resp.Diagnostics.Append(diag...)

	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	data.SecretProvider = types.StringValue(openSecret.Provider)
	data.SecretValues = secretsOutput

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
