// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packerconfig

import "github.com/hashicorp/terraform-provider-hcp/internal/provider/packer/testutils/configbuilder"

type VersionDataSourceBuilder interface {
	configbuilder.DataSourceBuilder

	SetBucketName(bucketName string)
	GetBucketName() string
	SetChannelName(channelName string)
	GetChannelName() string
}

func NewVersionDataSourceBuilder(uniqueName string) VersionDataSourceBuilder {
	return &versionDataSourceBuilder{
		newPackerDataSourceBuilder("version", uniqueName),
	}
}

func CloneVersionDataSourceBuilder(oldBuilder VersionDataSourceBuilder) VersionDataSourceBuilder {
	return &versionDataSourceBuilder{
		configbuilder.CloneDataSourceBuilder(oldBuilder),
	}
}

type versionDataSourceBuilder struct {
	configbuilder.DataSourceBuilder
}

var _ VersionDataSourceBuilder = &versionDataSourceBuilder{}

func (b *versionDataSourceBuilder) SetBucketName(bucketName string) {
	b.SetAttribute("bucket_name", bucketName)
}

func (b *versionDataSourceBuilder) GetBucketName() string {
	return b.GetAttribute("bucket_name")
}

func (b *versionDataSourceBuilder) SetChannelName(channelName string) {
	b.SetAttribute("channel_name", channelName)
}

func (b *versionDataSourceBuilder) GetChannelName() string {
	return b.GetAttribute("channel_name")
}
