// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

// OrganizationLocation represents the location of a target in HCP,
// identified by an organization ID.
//
// For example, a Project is located at an OrganizationLocation.
type OrganizationLocation interface {
	GetOrganizationID() string
}

// ProjectLocation represents the location of a target in HCP,
// identified by an organization ID and a project ID.
//
// For example, a Bucket is located at a ProjectLocation.
type ProjectLocation interface {
	OrganizationLocation
	GetProjectID() string
}

type Location = ProjectLocation

type GenericLocation struct {
	OrganizationID string
	ProjectID      string
}

var _ ProjectLocation = GenericLocation{}

func (l GenericLocation) GetOrganizationID() string {
	return l.OrganizationID
}

func (l GenericLocation) GetProjectID() string {
	return l.ProjectID
}

// BucketLocation represents the location of a target in HCP,
// identified by an organization ID, a project ID, and a bucket name.
//
// For example, a Channel is located at a BucketLocation.
type BucketLocation interface {
	Location
	GetBucketName() string
}

type GenericBucketLocation struct {
	Location
	BucketName string
}

var _ BucketLocation = GenericBucketLocation{}

func (l GenericBucketLocation) GetBucketName() string {
	return l.BucketName
}

// ChannelLocation represents the location of a target in HCP,
// identified by an organization ID, a project ID, a bucket name, and a channel name.
//
// For example, a Version is located at a ChannelLocation.
type ChannelLocation interface {
	BucketLocation
	GetChannelName() string
}

type GenericChannelLocation struct {
	BucketLocation
	ChannelName string
}

var _ ChannelLocation = GenericChannelLocation{}

func (l GenericChannelLocation) GetChannelName() string {
	return l.ChannelName
}

// VersionLocation represents the location of a target in HCP,
// identified by an organization ID, a project ID, a bucket name, and a version fingerprint.
//
// For example, an Artifact is located at a VersionLocation.
type VersionLocation interface {
	BucketLocation
	GetVersionFingerprint() string
}

type GenericVersionLocation struct {
	BucketLocation
	VersionFingerprint string
}

var _ VersionLocation = GenericVersionLocation{}

func (l GenericVersionLocation) GetVersionFingerprint() string {
	return l.VersionFingerprint
}
