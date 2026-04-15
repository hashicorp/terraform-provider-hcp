// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package packerv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"

	buildservice "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2022-12-02/client/build_service"
	models20221202 "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2022-12-02/models"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// GetImageByBuildLabelsResult holds the version, build, and artifact returned by GetImageByBuildLabels.
// The artifact is the one matching the requested platform/region (or the single artifact if the API filtered).
type GetImageByBuildLabelsResult struct {
	Version  *Version
	Build    *Build
	Artifact *Artifact
}

// GetImageByBuildLabels calls the HCP Packer GetImageByBuildLabels API (2022-12-02) to resolve
// the single image from the channel's current iteration whose build labels match the request.
func GetImageByBuildLabels(
	ctx context.Context,
	client *clients.Client,
	loc location.BucketLocation,
	bucketName, channelName string,
	labels map[string]string,
	cloudProvider, region string,
) (*GetImageByBuildLabelsResult, error) {
	if client.PackerBuildService == nil {
		return nil, fmt.Errorf("build label filtering is not available: PackerBuildService client is not configured")
	}
	params := buildservice.NewGetImageByBuildLabelsParamsWithContext(ctx)
	params.SetBucketName(bucketName)
	params.SetChannelName(channelName)
	params.SetLocationOrganizationID(loc.GetOrganizationID())
	params.SetLocationProjectID(loc.GetProjectID())
	params.SetBody(&models20221202.GetImageByBuildLabelsRequestBody{
		Labels:        labels,
		CloudProvider: cloudProvider,
		Region:        region,
	})
	resp, err := client.PackerBuildService.GetImageByBuildLabels(params, nil)
	if err != nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: %w", err)
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Artifact == nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: empty response")
	}
	ext := resp.Payload.Artifact
	result, err := mapExternalArtifactToResult(ext, bucketName, region)
	if err != nil {
		return nil, err
	}
	// Nested artifacts in the GetImageByBuildLabels payload omit created_at; fill from GetVersion
	// so the data source matches channel/fingerprint lookups.
	enrichArtifactFromFullVersion(client, loc, result, region)
	return result, nil
}

// enrichArtifactFromFullVersion replaces result.Artifact with the same artifact from GetVersionByFingerprint
// when that response includes CreatedAt (and other fields) missing from the build-labels API shape.
func enrichArtifactFromFullVersion(client *clients.Client, loc location.BucketLocation, result *GetImageByBuildLabelsResult, wantRegion string) {
	if client == nil || result == nil || result.Version == nil || result.Build == nil || result.Artifact == nil {
		return
	}
	fp := result.Version.Fingerprint
	if fp == "" {
		return
	}
	fullVer, err := GetVersionByFingerprint(client, loc, fp)
	if err != nil || fullVer == nil {
		return
	}
	buildID := result.Build.ID
	artifactID := result.Artifact.ID
	for _, b := range fullVer.Builds {
		if b == nil || b.ID != buildID {
			continue
		}
		for _, a := range b.Artifacts {
			if a == nil {
				continue
			}
			if a.ID == artifactID && (wantRegion == "" || a.Region == wantRegion) {
				if !a.CreatedAt.IsZero() {
					result.Artifact = a
				}
				return
			}
		}
		for _, a := range b.Artifacts {
			if a != nil && wantRegion != "" && a.Region == wantRegion && !a.CreatedAt.IsZero() {
				result.Artifact = a
				return
			}
		}
		return
	}
}

func mapExternalArtifactToResult(ext *models20221202.ExternalArtifact, bucketName, wantRegion string) (*GetImageByBuildLabelsResult, error) {
	if ext.Build == nil {
		return nil, fmt.Errorf("GetImageByBuildLabels response missing build")
	}
	// Map version
	version := &packermodels.HashicorpCloudPacker20230101Version{
		BucketName: bucketName,
	}
	if ext.Version != nil {
		version.Fingerprint = ext.Version.Fingerprint
		version.ID = ext.Version.ID
		if !ext.Version.RevokeAt.IsZero() {
			version.RevokeAt = ext.Version.RevokeAt
		}
	}
	if ext.Bucket != nil && version.BucketName == "" {
		version.BucketName = ext.Bucket.Name
	}
	// Map build
	b := ext.Build
	build := &packermodels.HashicorpCloudPacker20230101Build{
		ID:            b.ID,
		VersionID:     b.VersionID,
		ComponentType: b.ComponentType,
		PackerRunUUID: b.PackerRunUUID,
		Platform:      b.Platform,
		Labels:        b.Labels,
		Artifacts:     make([]*packermodels.HashicorpCloudPacker20230101Artifact, 0, len(b.Artifacts)),
	}
	for _, a := range b.Artifacts {
		if a == nil {
			continue
		}
		build.Artifacts = append(build.Artifacts, &packermodels.HashicorpCloudPacker20230101Artifact{
			ID:                 a.ID,
			ExternalIdentifier: a.ExternalIdentifier,
			Region:             a.Region,
		})
	}
	// Pick artifact by region
	var artifact *packermodels.HashicorpCloudPacker20230101Artifact
	for _, a := range build.Artifacts {
		if wantRegion != "" && a.Region != wantRegion {
			continue
		}
		artifact = a
		break
	}
	if artifact == nil && len(build.Artifacts) > 0 {
		artifact = build.Artifacts[0]
	}
	if artifact == nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: no artifact found for region %q", wantRegion)
	}
	return &GetImageByBuildLabelsResult{
		Version:  version,
		Build:    build,
		Artifact: artifact,
	}, nil
}

// GetImageByBuildLabelsDiags is the diag-returning variant for use in the artifact data source.
func GetImageByBuildLabelsDiags(
	ctx context.Context,
	client *clients.Client,
	loc location.BucketLocation,
	bucketName, channelName string,
	labels map[string]string,
	cloudProvider, region string,
) (*GetImageByBuildLabelsResult, diag.Diagnostics) {
	var diags diag.Diagnostics
	result, err := GetImageByBuildLabels(ctx, client, loc, bucketName, channelName, labels, cloudProvider, region)
	if err != nil {
		diags.AddError(
			"failed to get image by build labels",
			fmt.Sprintf("GetImageByBuildLabels: %s", err.Error()),
		)
		return nil, diags
	}
	return result, diags
}
