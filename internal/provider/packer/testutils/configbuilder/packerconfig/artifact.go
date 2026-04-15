// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder"
)

type ArtifactDataSourceBuilder interface {
	configbuilder.DataSourceBuilder

	SetBucketName(bucketName string)
	GetBucketName() string
	SetChannelName(channelName string)
	GetChannelName() string
	SetVersionFingerprint(versionFingerprint string)
	GetVersionFingerprint() string
	SetPlatform(platform string)
	GetPlatform() string
	SetRegion(region string)
	GetRegion() string
	SetComponentType(componentType string)
	GetComponentType() string
	SetLabels(labels map[string]string)
	GetLabels() string
}

func NewArtifactDataSourceBuilder(uniqueName string) ArtifactDataSourceBuilder {
	return &artifactDataSourceBuilder{
		newPackerDataSourceBuilder("artifact", uniqueName),
	}
}

func CloneArtifactDataSourceBuilder(oldBuilder ArtifactDataSourceBuilder) ArtifactDataSourceBuilder {
	return &artifactDataSourceBuilder{
		configbuilder.CloneDataSourceBuilder(oldBuilder),
	}
}

type artifactDataSourceBuilder struct {
	configbuilder.DataSourceBuilder
}

var _ ArtifactDataSourceBuilder = &artifactDataSourceBuilder{}

func (b *artifactDataSourceBuilder) SetBucketName(bucketName string) {
	b.SetAttribute("bucket_name", bucketName)
}

func (b *artifactDataSourceBuilder) GetBucketName() string {
	return b.GetAttribute("bucket_name")
}

func (b *artifactDataSourceBuilder) SetChannelName(channelName string) {
	b.SetAttribute("channel_name", channelName)
}

func (b *artifactDataSourceBuilder) GetChannelName() string {
	return b.GetAttribute("channel_name")
}

func (b *artifactDataSourceBuilder) SetVersionFingerprint(versionFingerprint string) {
	b.SetAttribute("version_fingerprint", versionFingerprint)
}

func (b *artifactDataSourceBuilder) GetVersionFingerprint() string {
	return b.GetAttribute("version_fingerprint")
}

func (b *artifactDataSourceBuilder) SetPlatform(platform string) {
	b.SetAttribute("platform", platform)
}

func (b *artifactDataSourceBuilder) GetPlatform() string {
	return b.GetAttribute("platform")
}

func (b *artifactDataSourceBuilder) SetRegion(region string) {
	b.SetAttribute("region", region)
}

func (b *artifactDataSourceBuilder) GetRegion() string {
	return b.GetAttribute("region")
}

func (b *artifactDataSourceBuilder) SetComponentType(componentType string) {
	b.SetAttribute("component_type", componentType)
}

func (b *artifactDataSourceBuilder) GetComponentType() string {
	return b.GetAttribute("component_type")
}

// SetLabels sets the labels attribute for build-label filtering. The map is formatted as HCL.
// When set with channel_name, the data source uses the GetImageByBuildLabels (by labels) path.
func (b *artifactDataSourceBuilder) SetLabels(labels map[string]string) {
	b.SetAttribute("labels", labelsToHCL(labels))
}

func (b *artifactDataSourceBuilder) GetLabels() string {
	return b.GetAttribute("labels")
}

// labelsToHCL formats a map as Terraform HCL map literal.
func labelsToHCL(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range labels {
		sb.WriteString(fmt.Sprintf("  %q = %q\n", k, v))
	}
	sb.WriteString("}")
	return sb.String()
}
