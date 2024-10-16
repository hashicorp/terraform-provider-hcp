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

type DataSourceVaultSecretsApp struct {
	client *clients.Client
}

type DataSourceVaultSecretsAppModel struct {
	ID        types.String `tfsdk:"id"`
	AppName   types.String `tfsdk:"app_name"`
	ProjectID types.String `tfsdk:"project_id"`
	OrgID     types.String `tfsdk:"organization_id"`
	Secrets   types.Map    `tfsdk:"secrets"`
}

func NewVaultSecretsAppDataSource() datasource.DataSource {
	return &DataSourceVaultSecretsApp{}
}

func (d *DataSourceVaultSecretsApp) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_secrets_app"
}

func (d *DataSourceVaultSecretsApp) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Vault Secrets app data source retrieves secrets and their latest version values for a given application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
			"app_name": schema.StringAttribute{
				Description: "The name of the Vault Secrets application.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Vault Secrets app is located.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Vault Secrets app is located.",
				Computed:    true,
			},
			"secrets": schema.MapAttribute{
				Description: "A map of all secrets in the Vault Secrets app. Key is the secret name, value is the latest secret version value.",
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *DataSourceVaultSecretsApp) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceVaultSecretsApp) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceVaultSecretsAppModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	client := d.client
	if d.client == nil {
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

	appSecrets, err := clients.OpenVaultSecretsAppSecrets(ctx, client, loc, data.AppName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	openAppSecrets := map[string]string{}
	for _, appSecret := range appSecrets {
		switch {
		case appSecret.StaticVersion != nil:
			openAppSecrets[appSecret.Name] = appSecret.StaticVersion.Value
		case appSecret.RotatingVersion != nil:
			for name, value := range appSecret.RotatingVersion.Values {
				openAppSecrets[appSecret.Name+"_"+name] = value
			}
		case appSecret.DynamicInstance != nil:
			for name, value := range appSecret.DynamicInstance.Values {
				openAppSecrets[appSecret.Name+"_"+name] = value
			}
		default:
			resp.Diagnostics.AddError(
				"Unsupported HCP Secret type",
				fmt.Sprintf("HCP Secrets secret type %q is not currently supported by terraform-provider-hcp", appSecret.Type),
			)
			return
		}
	}

	data.ID = data.AppName
	data.OrgID = types.StringValue(client.Config.OrganizationID)
	data.ProjectID = types.StringValue(client.Config.ProjectID)
	secretsMap, diag := types.MapValueFrom(ctx, types.StringType, openAppSecrets)
	resp.Diagnostics.Append(diag...)
	data.Secrets = secretsMap
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
