// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package packerv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/utils/location"

	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// GetImageByBuildLabelsResult holds the version, build, and artifact returned by GetImageByBuildLabels.
// The artifact is the one matching the requested platform/region (or the single artifact if the API filtered).
type GetImageByBuildLabelsResult struct {
	Version  *Version
	Build    *Build
	Artifact *Artifact
}

// GetImageByBuildLabels resolves the single artifact from the channel's current version whose
// build labels match the request. It uses GetVersionByChannelName (stable/2023-01-01) and
// performs label matching client-side, replacing the removed stable/2022-12-02 API.
func GetImageByBuildLabels(
	ctx context.Context,
	client *clients.Client,
	loc location.BucketLocation,
	bucketName, channelName string,
	labels map[string]string,
	cloudProvider, region string,
) (*GetImageByBuildLabelsResult, error) {
	version, err := GetVersionByChannelName(client, loc, channelName)
	if err != nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: %w", err)
	}
	if version == nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: channel %q has no assigned version", channelName)
	}

	build, artifact := findBuildAndArtifact(version.Builds, labels, cloudProvider, region)
	if build == nil || artifact == nil {
		return nil, fmt.Errorf("GetImageByBuildLabels: no artifact found matching labels %v for region %q in channel %q", labels, region, channelName)
	}

	return &GetImageByBuildLabelsResult{
		Version:  version,
		Build:    build,
		Artifact: artifact,
	}, nil
}

// findBuildAndArtifact iterates builds to find one whose labels are a superset of the requested
// labels and whose platform matches cloudProvider (if non-empty), then picks the artifact for
// the requested region.
func findBuildAndArtifact(
	builds []*packermodels.HashicorpCloudPacker20230101Build,
	labels map[string]string,
	cloudProvider, region string,
) (*Build, *Artifact) {
	for _, b := range builds {
		if b == nil {
			continue
		}
		if cloudProvider != "" && b.Platform != cloudProvider {
			continue
		}
		if !labelsMatch(b.Labels, labels) {
			continue
		}
		var fallback *Artifact
		for _, a := range b.Artifacts {
			if a == nil {
				continue
			}
			if region != "" && a.Region == region {
				return b, a
			}
			if fallback == nil {
				fallback = a
			}
		}
		if fallback != nil {
			return b, fallback
		}
	}
	return nil, nil
}

// labelsMatch returns true when buildLabels contains every key-value pair in wantLabels.
func labelsMatch(buildLabels, wantLabels map[string]string) bool {
	for k, v := range wantLabels {
		if buildLabels[k] != v {
			return false
		}
	}
	return true
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
