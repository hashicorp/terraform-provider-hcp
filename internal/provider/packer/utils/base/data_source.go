// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package base

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/customtypes"
)

type DataSourceParams struct {
	// An prefix for the data source type (Optional)
	// Typically used for the HCP product name
	// Formatted with lowercase and underscores
	// Example: `"packer"` for `data.hcp_packer_version`
	TypeNamePrefix string
	// The data source type's identifier (Required)
	// Formatted with lowercase and underscores
	// Example: `"version"` for `data.hcp_packer_version` if TypeNamePrefix is `"packer"`
	TypeName string
	// The data source type's "pretty" name (Optional)
	// Formatted with title case and spaces
	// Used for templated error messages and descriptions of common schema elements
	// If not provided, the TypeName will be used instead
	// Example: `"Version"` for `data.hcp_packer_version` or `"Channel Assignment"` for `hcp_packer_channel_assignment`
	PrettyName string
	// The data source schema (Required)
	// Additional common schema elements will be injected by `NewDataSourceBase`
	Schema schema.Schema
}

// NewPackerDataSource creates a new data source with common attributes injected
//
// If TypeNamePrefix is provided, it will be suffixed with `packer_`
// If TypeNamePrefix is not provided, it will be set to `packer`
func NewPackerDataSource(params DataSourceParams) DataSourceBase {
	if params.PrettyName == "" {
		params.PrettyName = params.TypeName
	}

	// Update the TypeNamePrefix to start with "packer" if it does not already
	if params.TypeNamePrefix == "" {
		params.TypeNamePrefix = "packer"
	}
	if params.TypeNamePrefix != "packer" {
		params.TypeNamePrefix = fmt.Sprintf("%s_%s", "packer", params.TypeNamePrefix)
	}

	params.Schema.Attributes["organization_id"] = &schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: fmt.Sprintf("The ID of the HCP Organization where the %s is located", params.PrettyName),
		Computed:    true,
	}
	params.Schema.Attributes["project_id"] = &schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: fmt.Sprintf("The ID of the HCP Organization where the %s is located", params.PrettyName),
		Optional:    true,
		Computed:    true,
	}

	return newDataSource(params)
}

func newDataSource(params DataSourceParams) DataSourceBase {
	return &dataSource{
		TypeNamePrefix: params.TypeNamePrefix,
		TypeName:       params.TypeName,
		schema:         params.Schema,
	}
}

type DataSourceBase interface {
	Metadata(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
	Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse)
	Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse)
	Client() *clients.Client
}

func NewDataSourceConfigValidatorMixin(validators ...datasource.ConfigValidator) DataSourceConfigValidatorMixin {
	return &dataSourceConfigValidatorMixin{
		validators: validators,
	}
}

type DataSourceConfigValidatorMixin interface {
	ConfigValidators(context.Context) []datasource.ConfigValidator
}

type dataSource struct {
	// An optional prefix for the data source type (ex. `packer` for `data.hcp_packer_version`)
	TypeNamePrefix string
	// The data source type (ex. `version` for `data.hcp_packer_version`) if the TypeNamePrefix is `packer`
	TypeName string
	schema   schema.Schema

	client *clients.Client
}

var _ DataSourceBase = &dataSource{}

func (ds *dataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if ds.TypeNamePrefix != "" {
		resp.TypeName = fmt.Sprintf("%s_%s_%s", req.ProviderTypeName, ds.TypeNamePrefix, ds.TypeName)
		return
	}

	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, ds.TypeName)
}

func (ds *dataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	ds.client = client
}

func (ds *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ds.schema
}

func (ds *dataSource) Client() *clients.Client {
	return ds.client
}

type dataSourceConfigValidatorMixin struct {
	validators []datasource.ConfigValidator
}

func (dscvm *dataSourceConfigValidatorMixin) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return dscvm.validators
}
