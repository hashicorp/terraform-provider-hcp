package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
)

const (
	// ConsulClusterResourceType is the resource type of a Consul cluster
	ConsulClusterResourceType = "hashicorp.consul.cluster"

	// HvnResourceType is the resource type of an HVN
	HvnResourceType = "hashicorp.network.hvn"

	// PeeringResourceType is the resource type of a network peering
	PeeringResourceType = "hashicorp.network.peering"

	// ConsulSnapshotResourceType is the resource type of a Consul snapshot
	ConsulSnapshotResourceType = "hashicorp.consul.snapshot"
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
// The resulting link location does not include an organization, which is
// typically required for requests.
func parseLinkURL(urn string, resourceType string) (*sharedmodels.HashicorpCloudLocationLink, error) {
	pattern := fmt.Sprintf("^/project/[^/]+/%s/[^/]+$", resourceType)
	match, _ := regexp.MatchString(pattern, urn)
	if !match {
		return nil, fmt.Errorf("url is not in the correct format: /project/{project_id}/%s/{id}", resourceType)
	}

	components := strings.Split(urn, "/")

	return &sharedmodels.HashicorpCloudLocationLink{
		Type: components[3],
		ID:   components[4],
		Location: &sharedmodels.HashicorpCloudLocationLocation{
			ProjectID: components[2],
		},
	}, nil
}
