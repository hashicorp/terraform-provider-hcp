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
				// Optional: when set with channel_name, resolves the artifact by build labels (GetImageByBuildLabels API)
				"labels": schema.MapAttribute{
					ElementType: basetypes.StringType{},
					Description: "Labels associated with the build. When set with `channel_name`, the artifact is resolved by matching these build labels (e.g. `{ \"nomad_version\" = \"1.8.10\" }`). When unset, computed from the resolved build.",
					MarkdownDescription: "When set (non-empty) with `channel_name`, the data source uses **GetImageByBuildLabels**: the channel's current version is scanned for builds whose labels contain every key/value you supply.\n\n" +
						"If more than one build matches, the HCP Packer API returns the **first** candidate in **`updated_at` descending** order that also matches the **`platform`** and **`region`** you configure (sent as cloud provider and region). Always set `platform` and `region` so the result is predictable when multiple builds or clouds exist.\n\n" +
						"If several builds still match the same platform and region (for example different Packer sources), use **`component_type`** to disambiguate, consistent with the non-label lookup path.\n\n" +
						"When `labels` is unset or empty, this attribute is computed from the resolved build.",
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
			labelsRequireChannelValidator{},
		),
	}
}

// labelsRequireChannelValidator requires channel_name when labels is set (label filtering only works with channel).
type labelsRequireChannelValidator struct{}

func (labelsRequireChannelValidator) Description(ctx context.Context) string {
	return "When labels is set, channel_name must also be set."
}

func (labelsRequireChannelValidator) MarkdownDescription(ctx context.Context) string {
	return "When `labels` is set, `channel_name` must also be set. Build label filtering is only supported when resolving by channel."
}

func (labelsRequireChannelValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var model dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model.Labels.IsNull() || model.Labels.IsUnknown() {
		return
	}
	labelsMap := labelsMapFromValue(model.Labels)
	if len(labelsMap) == 0 {
		return
	}
	if model.ChannelName.IsNull() || model.ChannelName.IsUnknown() || model.ChannelName.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid combination of arguments",
			"When `labels` is set, `channel_name` must also be set. Build label filtering is only supported when resolving by channel.",
		)
	}
}

// labelsMapFromValue extracts map[string]string from a MapValue (element type string).
func labelsMapFromValue(m basetypes.MapValue) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := make(map[string]string)
	for k, v := range m.Elements() {
		if s, ok := v.(basetypes.StringValue); ok && !s.IsNull() && !s.IsUnknown() {
			out[k] = s.ValueString()
		}
	}
	return out
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

	// When labels is set with channel_name, use GetImageByBuildLabels API instead of GetChannel + filter.
	useLabelsFilter := !model.Labels.IsNull() && !model.Labels.IsUnknown() && len(labelsMapFromValue(model.Labels)) > 0
	if useLabelsFilter && !model.ChannelName.IsNull() && !model.ChannelName.IsUnknown() {
		result, getDiags := packerv2.GetImageByBuildLabelsDiags(
			ctx,
			d.Client(),
			model,
			model.BucketName.ValueString(),
			model.ChannelName.ValueString(),
			labelsMapFromValue(model.Labels),
			model.Platform.ValueString(),
			model.Region.ValueString(),
		)
		resp.Diagnostics.Append(getDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if result == nil || result.Version == nil || result.Build == nil || result.Artifact == nil {
			resp.Diagnostics.AddError(
				"HCP Packer Artifact not found",
				"No artifact matched the given labels and filters (channel_name, platform, region).",
			)
			return
		}
		model.populateFromVersion(result.Version)
		resp.Diagnostics.Append(model.populateFromBuild(result.Build)...)
		resp.Diagnostics.Append(model.populateFromArtifact(result.Artifact)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

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
