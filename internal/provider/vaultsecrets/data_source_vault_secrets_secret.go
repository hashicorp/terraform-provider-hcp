// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"context"
	"encoding/json"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

type DataSourceVaultSecretsSecret struct {
	client *clients.Client
}

type DataSourceVaultSecretsSecretModel struct {
	ID          types.String `tfsdk:"id"`
	AppName     types.String `tfsdk:"app_name"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
	SecretName  types.String `tfsdk:"secret_name"`
	SecretValue types.String `tfsdk:"secret_value"`
}

func NewVaultSecretsSecretDataSource() datasource.DataSource {
	return &DataSourceVaultSecretsSecret{}
}

func (d *DataSourceVaultSecretsSecret) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_secret"
}

func (d *DataSourceVaultSecretsSecret) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets secret data source retrieves a singular secret and its latest version.",
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

func (d *DataSourceVaultSecretsSecret) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceVaultSecretsSecret) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceVaultSecretsSecretModel
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

	// NOTE: for backwards compatibility purposes, if the secret is not a static secret (a string)
	// encode the complex secret as a JSON string
	var secretValue string
	switch {
	case openSecret.StaticVersion != nil:
		secretValue = openSecret.StaticVersion.Value
	case openSecret.RotatingVersion != nil:
		secretData, err := json.Marshal(openSecret.RotatingVersion.Values)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "could not encode rotating secret as json")
			return
		}
		resp.Diagnostics.AddWarning(
			"HCP Vault Secrets mismatched type",
			"Attempted to get a rotating secret in a KV secret data source, encoding the secret values as JSON",
		)
		secretValue = string(secretData)
	default:
		resp.Diagnostics.AddError(
			"Unsupported HCP Secret type",
			fmt.Sprintf("HCP Secrets secret type %q is not currently supported by terraform-provider-hcp", openSecret.Type),
		)
		return
	}

	data.ID = data.AppName
	data.SecretValue = types.StringValue(secretValue)
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
