// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// NOTE: The `Link` behavior in this file is based off of the internal cloud-api:
// https://github.com/hashicorp/cloud-api-internal/blob/master/helper/hashicorp/cloud/location/link.go
//
// It is important that the implementation here is consistent with the internal
// cloud-api because the `Link`s produced by these functions could be sent in
// API requests. In practice, this primarily means that the resource types must
// be the same in both places, eg. the HVN type is defined here:
// https://github.com/hashicorp/cloud-network/blob/master/resource/network.go#L13

const (
	// ConsulClusterResourceType is the resource type of a Consul cluster
	ConsulClusterResourceType = "hashicorp.consul.cluster"

	// HvnResourceType is the resource type of an HVN
	HvnResourceType = "hashicorp.network.hvn"

	// PeeringResourceType is the resource type of a network peering
	PeeringResourceType = "hashicorp.network.peering"

	// TgwAttachmentResourceType is the resource type of a TGW attachment
	TgwAttachmentResourceType = "hashicorp.network.tgw-attachment"

	// HVNRouteResourceType is the resource type of an HVN route
	HVNRouteResourceType = "hashicorp.network.route"

	// ConsulSnapshotResourceType is the resource type of a Consul snapshot
	ConsulSnapshotResourceType = "hashicorp.consul.snapshot"

	// ConsulClusterHelmConfigDataSourceType is the data source type of a Consul
	// cluster Helm config
	ConsulClusterHelmConfigDataSourceType = ConsulClusterResourceType + ".helm-config"

	// ConsulClusterAgentKubernetesSecretDataSourceType is the data source
	// type of a Consul cluster agent Kubernetes secret
	ConsulClusterAgentKubernetesSecretDataSourceType = ConsulClusterResourceType + ".agent-kubernetes-secret"

	// VaultClusterResourceType is the resource type of a Vault cluster
	VaultClusterResourceType = "hashicorp.vault.cluster"

	// BoundaryClusterResourceType is the resource type of a Boundary Cluster
	BoundaryClusterResourceType = "hashicorp.boundary.cluster"
)

// newLink constructs a new Link from the passed arguments. ID should be the
// user specified resource ID.
//
// Adapted from https://github.com/hashicorp/cloud-api-internal/blob/master/helper/hashicorp/cloud/location/link.go#L10-L23
func newLink(loc *sharedmodels.HashicorpCloudLocationLocation, resourceType string, id string) *sharedmodels.HashicorpCloudLocationLink {
	return &sharedmodels.HashicorpCloudLocationLink{
		Type:     resourceType,
		ID:       id,
		Location: loc,
	}
}

// linkURL generates a URL from the passed link. If the link is invalid, an
// error is returned. The Link URL is a globally unique, human readable string
// identifying a resource.
// This version of the function includes org and project data, but not provider
// and region.
//
// Adapted from https://github.com/hashicorp/cloud-api-internal/blob/master/helper/hashicorp/cloud/location/link.go#L25-L60
func linkURL(l *sharedmodels.HashicorpCloudLocationLink) (string, error) {
	if l == nil {
		return "", errors.New("nil link")
	}

	if l.Location == nil {
		return "", errors.New("link missing Location")
	}

	// Validate that the link contains the necessary information
	if l.Location.ProjectID == "" {
		return "", errors.New("link missing project ID")
	} else if l.Type == "" {
		return "", errors.New("link missing resource type")
	}

	// Determine the ID of the resource
	id := l.ID
	if id == "" {
		return "", errors.New("link missing resource ID")
	}

	// Generate the URL
	urn := fmt.Sprintf("/project/%s/%s/%s",
		l.Location.ProjectID,
		l.Type,
		id)

	return urn, nil
}

// parseLinkURL parses a link URL into a link. If the URL is malformed, an
// error is returned.
//
// If `expectedType` is provided it will be matched against the resource from
// the URL and if they don't match the function returns an error. If `expectedType`
// is an empty string then the resource type just will be inferred from the URL
// as is.
//
// The resulting link location does not include an organization, which is
// typically required for requests. If organization is needed, use
// `buildLinkFromURL()`.
func parseLinkURL(urn string, expectedType string) (*sharedmodels.HashicorpCloudLocationLink, error) {
	pattern := "^/project/[^/]+/[^/]+/[^/]+$"
	match, _ := regexp.MatchString(pattern, urn)
	if !match {
		return nil, fmt.Errorf("url %q is not in the correct format: /project/{project_id}/{resource_type}/{id}", urn)
	}

	components := strings.Split(urn, "/")

	if expectedType != "" && expectedType != components[3] {
		return nil, fmt.Errorf("url %q is not in the correct format: /project/{project_id}/%s/{id}", urn, expectedType)
	}

	return &sharedmodels.HashicorpCloudLocationLink{
		Type: components[3],
		ID:   components[4],
		Location: &sharedmodels.HashicorpCloudLocationLocation{
			ProjectID: components[2],
		},
	}, nil
}

// buildLinkFromURL builds a full link from a link URL. In particular, a link
// URL only contains the project ID of its location, so this function populates
// the organization ID, which is required for most requests.
func buildLinkFromURL(urn string, resourceType string, organizationID string) (*sharedmodels.HashicorpCloudLocationLink, error) {
	link, err := parseLinkURL(urn, resourceType)
	if err != nil {
		return nil, err
	}

	link.Location.OrganizationID = organizationID

	return link, nil
}
