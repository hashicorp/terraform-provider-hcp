// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients/packerv2"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/customtypes"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/base"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"
)

func NewDataSource() datasource.DataSource {
	params := base.DataSourceParams{
		TypeName:   "version",
		PrettyName: "Version",
		Schema: schema.Schema{
			Description: "The HCP Packer Version data source retrieves information about a Version.",
			Attributes: map[string]schema.Attribute{
				// Required Inputs
				"bucket_name": schema.StringAttribute{
					CustomType:  customtypes.SlugType{},
					Description: "The name of the HCP Packer Bucket where the Version is located",
					Required:    true,
				},
				"channel_name": schema.StringAttribute{
					CustomType:  customtypes.SlugType{},
					Description: "The name of the HCP Packer Channel the Version is assigned to",
					MarkdownDescription: `
The name of the HCP Packer Channel the Version is assigned to.
The version currently assigned to the Channel will be fetched.`,
					Required: true,
				},
				// Computed Outputs
				"fingerprint": schema.StringAttribute{
					CustomType:  customtypes.PackerFingerprintType{},
					Description: "The fingerprint of the HCP Packer Version",
					Computed:    true,
				},
				"id": schema.StringAttribute{
					CustomType:  customtypes.ULIDType{},
					Description: "The ULID of the HCP Packer Version",
					Computed:    true,
				},
				"name": schema.StringAttribute{
					Description: "The name of the HCP Packer Version",
					Computed:    true,
				},
				"author_id": schema.StringAttribute{
					Description: "The name of the person who created this HCP Packer Version",
					Computed:    true,
				},
				"created_at": schema.StringAttribute{
					Description: "The creation time of this HCP Packer Version",
					Computed:    true,
				},
				"updated_at": schema.StringAttribute{
					Description: "The last time this HCP Packer Version was updated",
					Computed:    true,
				},
				"revoke_at": schema.StringAttribute{
					Description: "The revocation time of this HCP Packer Version. " +
						"This field will be null for any Version that has not been revoked or scheduled for revocation",
					Computed: true,
				},
			},
		},
	}

	return &dataSource{
		DataSourceBase: base.NewPackerDataSource(params),
	}
}

type dataSource struct {
	base.DataSourceBase
}

var _ datasource.DataSource = &dataSource{}

type dataSourceModel struct {
	ProjectID      customtypes.UUIDValue `tfsdk:"project_id"`
	OrganizationID customtypes.UUIDValue `tfsdk:"organization_id"`

	BucketName  customtypes.SlugValue `tfsdk:"bucket_name"`
	ChannelName customtypes.SlugValue `tfsdk:"channel_name"`

	Fingerprint customtypes.PackerFingerprintValue `tfsdk:"fingerprint"`
	ID          customtypes.ULIDValue              `tfsdk:"id"`
	Name        basetypes.StringValue              `tfsdk:"name"`
	AuthorID    basetypes.StringValue              `tfsdk:"author_id"`

	CreatedAt basetypes.StringValue `tfsdk:"created_at"`
	UpdatedAt basetypes.StringValue `tfsdk:"updated_at"`
	RevokeAt  basetypes.StringValue `tfsdk:"revoke_at"`
}

var _ location.BucketLocation = dataSourceModel{}

func (m dataSourceModel) GetOrganizationID() string {
	return m.OrganizationID.ValueString()
}

func (m dataSourceModel) GetProjectID() string {
	return m.ProjectID.ValueString()
}

func (m dataSourceModel) GetBucketName() string {
	return m.BucketName.ValueString()
}

func (m *dataSourceModel) populateFromLocationIfEmpty(location location.Location) {
	if m.ProjectID.IsNull() || m.ProjectID.ValueString() == "" {
		m.ProjectID = customtypes.NewUUIDValue(location.GetProjectID())
	}
	if m.OrganizationID.IsNull() || m.OrganizationID.ValueString() == "" {
		m.OrganizationID = customtypes.NewUUIDValue(location.GetOrganizationID())
	}
}

func (m *dataSourceModel) populateFromVersion(version *packerv2.Version) {
	if version == nil {
		version = &packerv2.Version{}
	}

	m.BucketName = customtypes.NewSlugValue(version.BucketName)
	m.Fingerprint = customtypes.NewPackerFingerprintValue(version.Fingerprint)
	m.ID = customtypes.NewULIDValue(version.ID)
	m.Name = types.StringValue(version.Name)
	m.AuthorID = types.StringValue(version.AuthorID)
	m.CreatedAt = types.StringValue(version.CreatedAt.String())
	m.UpdatedAt = types.StringValue(version.UpdatedAt.String())
	m.RevokeAt = types.StringValue(version.RevokeAt.String())
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get and validate config model from the request
	var model dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	// Get and validate client
	client := d.Client()
	resp.Diagnostics.Append(utils.CheckClient(client)...)

	// Check for errors from previous steps
	if resp.Diagnostics.HasError() {
		return
	}

	model.populateFromLocationIfEmpty(client)

	version, getVersionDiags := packerv2.GetVersionByChannelNameDiags(d.Client(), model, model.ChannelName.ValueString())
	resp.Diagnostics.Append(getVersionDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.populateFromVersion(version)

	// Set the state from the data source model and append any errors to the response
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
