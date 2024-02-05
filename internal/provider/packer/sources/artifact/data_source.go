// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package artifact

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		TypeName:   "artifact",
		PrettyName: "Artifact",
		Schema: schema.Schema{
			Description: "The HCP Packer Artifact data source retrieves information about an Artifact.",
			Attributes: map[string]schema.Attribute{
				// Required Inputs
				"bucket_name": schema.StringAttribute{
					CustomType:  customtypes.SlugType{},
					Description: "The name of the HCP Packer Bucket where the Artifact is located.",
					Required:    true,
				},
				"platform": schema.StringAttribute{
					Description: "Name of the platform where the HCP Packer Artifact is stored.",
					Required:    true,
				},
				"region": schema.StringAttribute{
					Description: "The Region where the HCP Packer Artifact is stored, if any.",
					Required:    true,
				},
				// Optional Inputs
				"channel_name": schema.StringAttribute{
					CustomType:  customtypes.SlugType{},
					Description: "The name of the HCP Packer Channel the Version containing this Artifact is assigned to",
					MarkdownDescription: `
The name of the HCP Packer Channel the Version containing this Artifact is assigned to.
The Version currently assigned to the Channel will be fetched. 
Exactly one of ` + "`channel_name` or `version_fingerprint`" + ` must be provided.`,
					Optional: true,
				},
				"version_fingerprint": schema.StringAttribute{
					CustomType:  customtypes.PackerFingerprintType{},
					Description: "The fingerprint of the HCP Packer Version where the Artifact is located",
					MarkdownDescription: `
The fingerprint of the HCP Packer Version where the Artifact is located. 
If provided in the config, it is used to fetch the Version.
Exactly one of ` + "`channel_name` or `version_fingerprint`" + ` must be provided.`,
					Optional: true,
					Computed: true,
				},
				"component_type": schema.StringAttribute{
					Description: "Name of the Packer builder that built this Artifact. Ex: `amazon-ebs.example`.",
					// TODO: Add input validation for component_type
					Optional: true,
					Computed: true,
				},
				// Computed Outputs
				"id": schema.StringAttribute{
					CustomType:  customtypes.ULIDType{},
					Description: "The ULID of the HCP Packer Artifact.",
					Computed:    true,
				},
				"build_id": schema.StringAttribute{
					CustomType:  customtypes.ULIDType{},
					Description: "The ULID of the HCP Packer Build where the Artifact is located.",
					Computed:    true,
				},
				"packer_run_uuid": schema.StringAttribute{
					CustomType:  customtypes.UUIDType{},
					Description: "The UUID of the build containing this image.",
					Computed:    true,
				},
				"external_identifier": schema.StringAttribute{
					Description: "An external identifier for the HCP Packer Artifact.",
					Computed:    true,
				},
				"labels": schema.MapAttribute{
					ElementType: basetypes.StringType{},
					Description: "Labels associated with the build containing this image.",
					Computed:    true,
				},
				"created_at": schema.StringAttribute{
					Description: "The creation time of this HCP Packer Artifact.",
					Computed:    true,
				},
				"revoke_at": schema.StringAttribute{
					Description: "The revocation time of the HCP Packer Version containing this Artifact. " +
						"This field will be null for any Version that has not been revoked or scheduled for revocation.",
					Computed: true,
				},
			},
		},
	}

	return &dataSource{
		DataSourceBase: base.NewPackerDataSource(params),
		DataSourceConfigValidatorMixin: base.NewDataSourceConfigValidatorMixin(
			datasourcevalidator.ExactlyOneOf(
				path.MatchRelative().AtName("channel_name"),
				path.MatchRelative().AtName("version_fingerprint"),
			),
		),
	}
}

type dataSource struct {
	base.DataSourceBase
	base.DataSourceConfigValidatorMixin
}

var _ datasource.DataSource = &dataSource{}
var _ datasource.DataSourceWithConfigValidators = &dataSource{}

type dataSourceModel struct {
	OrganizationID customtypes.UUIDValue `tfsdk:"organization_id"`
	ProjectID      customtypes.UUIDValue `tfsdk:"project_id"`

	BucketName customtypes.SlugValue `tfsdk:"bucket_name"`
	Platform   basetypes.StringValue `tfsdk:"platform"`
	Region     basetypes.StringValue `tfsdk:"region"`

	ChannelName        customtypes.SlugValue              `tfsdk:"channel_name"`
	VersionFingerprint customtypes.PackerFingerprintValue `tfsdk:"version_fingerprint"`

	ComponentType basetypes.StringValue `tfsdk:"component_type"`

	ID                 customtypes.ULIDValue `tfsdk:"id"`
	BuildID            customtypes.ULIDValue `tfsdk:"build_id"`
	PackerRunUUID      customtypes.UUIDValue `tfsdk:"packer_run_uuid"`
	ExternalIdentifier basetypes.StringValue `tfsdk:"external_identifier"`

	Labels basetypes.MapValue `tfsdk:"labels"`

	CreatedAt basetypes.StringValue `tfsdk:"created_at"`
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

func (m *dataSourceModel) populateFromVersion(version *packerv2.Version) diag.Diagnostics {
	var diags diag.Diagnostics

	if version == nil {
		version = &packerv2.Version{}
	}

	m.BucketName = customtypes.NewSlugValue(version.BucketName)

	m.VersionFingerprint = customtypes.NewPackerFingerprintValue(version.Fingerprint)

	if !version.RevokeAt.IsZero() {
		m.RevokeAt = basetypes.NewStringValue(version.RevokeAt.String())
	}

	return diags
}

func (m *dataSourceModel) populateFromBuild(build *packerv2.Build) diag.Diagnostics {
	var diags diag.Diagnostics

	if build == nil {
		build = &packerv2.Build{}
	}

	m.Platform = basetypes.NewStringValue(build.Platform)

	m.ComponentType = basetypes.NewStringValue(build.ComponentType)

	m.BuildID = customtypes.NewULIDValue(build.ID)
	m.PackerRunUUID = customtypes.NewUUIDValue(build.PackerRunUUID)

	labels := map[string]attr.Value{}
	for k, v := range build.Labels {
		labels[k] = basetypes.NewStringValue(v)
	}

	labelValueMap, newDiags := basetypes.NewMapValue(
		basetypes.StringType{},
		labels,
	)
	diags.Append(newDiags...)
	if diags.HasError() {
		return diags
	}
	m.Labels = labelValueMap

	return diags
}

func (m *dataSourceModel) populateFromArtifact(artifact *packerv2.Artifact) diag.Diagnostics {
	if artifact == nil {
		artifact = &packerv2.Artifact{}
	}

	m.Region = types.StringValue(artifact.Region)

	m.ID = customtypes.NewULIDValue(artifact.ID)
	m.ExternalIdentifier = types.StringValue(artifact.ExternalIdentifier)

	m.CreatedAt = types.StringValue(artifact.CreatedAt.String())

	return nil
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

	var getVersionDiags diag.Diagnostics
	var version *packerv2.Version
	if !model.VersionFingerprint.IsUnknown() && !model.VersionFingerprint.IsNull() {
		version, getVersionDiags = packerv2.GetVersionByFingerprintDiags(d.Client(), model, model.VersionFingerprint.ValueString())
	} else if !model.ChannelName.IsUnknown() && !model.ChannelName.IsNull() {
		version, getVersionDiags = packerv2.GetVersionByChannelNameDiags(d.Client(), model, model.ChannelName.ValueString())
	} // else: should never happen due to config validation requiring exactly one of the two
	resp.Diagnostics.Append(getVersionDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.populateFromVersion(version)

	var build *packerv2.Build
	var artifact *packerv2.Artifact
	for _, b := range version.Builds {
		if b.Platform != model.Platform.ValueString() {
			continue
		}
		if model.ComponentType.ValueString() != "" && b.ComponentType != model.ComponentType.ValueString() {
			continue
		}
		index := slices.IndexFunc(
			b.Artifacts,
			func(artifact *packerv2.Artifact) bool {
				return artifact.Region == model.Region.ValueString()
			},
		)
		if index >= 0 {
			build = b
			artifact = b.Artifacts[index]
			break
		}
	}
	if build == nil || artifact == nil {
		resp.Diagnostics.AddError(
			"HCP Packer Artifact not found",
			fmt.Sprintf(
				"Could not find an Artifact with attributes (region: %q cloud: %q, version_fingerprint: %q, component_type: %q).",
				model.Region.ValueString(),
				model.Platform.ValueString(),
				model.VersionFingerprint.ValueString(),
				model.ComponentType.ValueString(),
			),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	model.populateFromBuild(build)
	model.populateFromArtifact(artifact)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
