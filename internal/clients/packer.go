// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetPackerChannelBySlug queries the HCP Packer registry for the channel
// associated with the given channel name.
func GetPackerChannelBySlug(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string) (*packermodels.HashicorpCloudPackerChannel, error) {

	getParams := packer_service.NewPackerServiceGetChannelParamsWithContext(ctx)
	getParams.BucketSlug = bucketName
	getParams.Slug = channelName
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Packer.PackerServiceGetChannel(getParams, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceGetChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by GetPackerChannelBySlug. Got: %v", err)
	}

	return getResp.Payload.Channel, nil
}

// GetPackerChannelBySlugFromList queries the HCP Packer registry for the
// channel associated with the given channel name, using ListBucketChannels
func GetPackerChannelBySlugFromList(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string) (*packermodels.HashicorpCloudPackerChannel, error) {
	resp, err := ListPackerChannels(ctx, client, loc, bucketName)
	if err != nil {
		return nil, err
	}

	for _, channel := range resp.Channels {
		if channel.Slug == channelName {
			return channel, nil
		}
	}

	return nil, nil
}

// GetIterationFromID queries the HCP Packer registry for an existing bucket iteration using its ULID.
func GetIterationFromID(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, iterationID string) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := newGetIterationParams(ctx, loc, bucketSlug)
	params.IterationID = &iterationID
	return getIteration(client, params)
}

// GetIterationFromVersion queries the HCP Packer registry for an existing bucket iteration using its incremental version.
func GetIterationFromVersion(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, iterationIncrementalVersion int32) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := newGetIterationParams(ctx, loc, bucketSlug)
	params.IncrementalVersion = &iterationIncrementalVersion
	return getIteration(client, params)
}

// GetIterationFromFingerprint queries the HCP Packer registry for an existing bucket iteration using its fingerprint.
func GetIterationFromFingerprint(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, iterationFingerprint string) (*packermodels.HashicorpCloudPackerIteration, error) {
	params := newGetIterationParams(ctx, loc, bucketSlug)
	params.Fingerprint = &iterationFingerprint
	return getIteration(client, params)
}

func newGetIterationParams(ctx context.Context, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketslug string) *packer_service.PackerServiceGetIterationParams {
	params := packer_service.NewPackerServiceGetIterationParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketslug
	return params
}

func getIteration(client *Client, params *packer_service.PackerServiceGetIterationParams) (*packermodels.HashicorpCloudPackerIteration, error) {
	it, err := client.Packer.PackerServiceGetIteration(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceGetIterationDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by getIteration. Got: %v", err)
	}

	return it.Payload.Iteration, nil
}

// CreatePackerChannel creates a channel on the named bucket.
func CreatePackerChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string, channelSlug string,
	restriction *packermodels.HashicorpCloudPackerCreateChannelRequestRestriction) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceCreateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Body.Slug = channelSlug
	params.Body.Restriction = restriction

	channel, err := client.Packer.PackerServiceCreateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceCreateChannelDefault); ok {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected error format received by CreateBucketChannel. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}

// UpdatePackerChannel updates the named channel.
func UpdatePackerChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string, channelSlug string,
	restricted bool) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceUpdateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelSlug
	params.Body.Mask = "restriction"

	if restricted {
		params.Body.Restriction = packermodels.HashicorpCloudPackerUpdateChannelRequestRestrictionRESTRICTED.Pointer()
	} else {
		params.Body.Restriction = packermodels.HashicorpCloudPackerUpdateChannelRequestRestrictionUNRESTRICTED.Pointer()
	}

	channel, err := client.Packer.PackerServiceUpdateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceUpdateChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by UpdateBucketChannel. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}

func UpdatePackerChannelAssignment(
	ctx context.Context, client *Client,
	loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketSlug string, channelSlug string,
	iterationFingerprint string,
) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceUpdateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelSlug

	maskPaths := []string{}

	if iterationFingerprint != "" {
		params.Body.Fingerprint = iterationFingerprint
		maskPaths = append(maskPaths, "fingerprint")
	}

	if len(maskPaths) == 0 {
		maskPaths = []string{"iterationId", "fingerprint", "incrementalVersion"}
	}
	params.Body.Mask = strings.Join(maskPaths, ",")

	channel, err := client.Packer.PackerServiceUpdateChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceUpdateChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by UpdateBucketChannel. Got: %v", err)
	}

	return channel.GetPayload().Channel, nil
}

// DeletePackerChannel deletes a channel from the named bucket.
func DeletePackerChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug, channelSlug string) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceDeleteChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelSlug

	req, err := client.Packer.PackerServiceDeleteChannel(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceDeleteChannelDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by DeleteBucketChannel. Got: %v", err)
	}

	if !req.IsSuccess() {
		return nil, errors.New("failed to delete channel")
	}

	return nil, nil
}

// ListPackerChannels queries the HCP Packer registry for channels associated to the specified bucket.
func ListPackerChannels(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string) (*packermodels.HashicorpCloudPackerListChannelsResponse, error) {
	params := packer_service.NewPackerServiceListChannelsParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug

	req, err := client.Packer.PackerServiceListChannels(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceListChannelsDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by ListBucketChannels. Got: %v", err)
	}

	return req.Payload, nil
}

// ListBuckets queries the HCP Packer registry for all associated buckets.
func ListBuckets(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation) ([]*packermodels.HashicorpCloudPackerBucket, error) {
	nextPage := ""
	var buckets []*packermodels.HashicorpCloudPackerBucket

	for {
		params := packer_service.NewPackerServiceListBucketsParams()
		params.LocationOrganizationID = loc.OrganizationID
		params.LocationProjectID = loc.ProjectID
		// Sort order is needed for acceptance tests.
		params.SortingOrderBy = []string{"slug"}
		if nextPage != "" {
			params.PaginationNextPageToken = &nextPage
		}

		req, err := client.Packer.PackerServiceListBuckets(params, nil)
		if err != nil {
			if err, ok := err.(*packer_service.PackerServiceListBucketsDefault); ok {
				return nil, errors.New(err.Payload.Message)
			}
			return nil, fmt.Errorf("unexpected error format received by ListBuckets. Got: %v", err)
		}

		buckets = append(buckets, req.Payload.Buckets...)
		pagination := req.Payload.Pagination
		if pagination == nil || pagination.NextPageToken == "" {
			return buckets, nil
		}

		nextPage = pagination.NextPageToken
	}
}

// GetRunTask queries the HCP Packer Registry for the API information needed to configure a run task
func GetRunTask(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation) (*packermodels.HashicorpCloudPackerGetRegistryTFCRunTaskAPIResponse, error) {
	params := packer_service.NewPackerServiceGetRegistryTFCRunTaskAPIParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.TaskType = "validation"

	req, err := client.Packer.PackerServiceGetRegistryTFCRunTaskAPI(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceGetRegistryTFCRunTaskAPIDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by GetRunTask. Got: %v", err)
	}

	return req.Payload, nil
}

// RegenerateHMAC triggers the HCP Packer Registry's run task HMAC Key to be regenerated
func RegenerateHMAC(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation) (*packermodels.HashicorpCloudPackerRegenerateTFCRunTaskHmacKeyResponse, error) {
	params := packer_service.NewPackerServiceRegenerateTFCRunTaskHmacKeyParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID

	req, err := client.Packer.PackerServiceRegenerateTFCRunTaskHmacKey(params, nil)
	if err != nil {
		if err, ok := err.(*packer_service.PackerServiceRegenerateTFCRunTaskHmacKeyDefault); ok {
			return nil, errors.New(err.Payload.Message)
		}
		return nil, fmt.Errorf("unexpected error format received by RegenerateHMAC. Got: %v", err)
	}

	return req.Payload, nil
}
