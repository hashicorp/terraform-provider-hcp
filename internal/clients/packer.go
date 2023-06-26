// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"errors"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/client/packer_service"
	packermodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
)

// GetPackerChannelBySlug queries the HCP Packer registry for the channel
// associated with the given channel name.
func GetPackerChannelBySlug(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string) (*packermodels.HashicorpCloudPackerChannel, error) {

	getParams := packer_service.NewPackerServiceGetChannelParams()
	getParams.BucketSlug = bucketName
	getParams.Slug = channelName
	getParams.LocationOrganizationID = loc.OrganizationID
	getParams.LocationProjectID = loc.ProjectID

	getResp, err := client.Packer.PackerServiceGetChannel(getParams, nil)
	if err != nil {
		return nil, handleGetChannelError(err.(*packer_service.PackerServiceGetChannelDefault))
	}

	return getResp.Payload.Channel, nil
}

// GetPackerChannelBySlugFromList queries the HCP Packer registry for the
// channel associated with the given channel name, using ListBucketChannels
func GetPackerChannelBySlugFromList(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation,
	bucketName string, channelName string) (*packermodels.HashicorpCloudPackerChannel, error) {
	resp, err := ListBucketChannels(ctx, client, loc, bucketName)
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
		return nil, handleGetIterationError(err.(*packer_service.PackerServiceGetIterationDefault))
	}

	return it.Payload.Iteration, nil
}

// CreateBucketChannel creates a channel on the named bucket.
func CreateBucketChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string, channelSlug string,
	iteration *packermodels.HashicorpCloudPackerIteration, restriction *packermodels.HashicorpCloudPackerCreateChannelRequestRestriction) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceCreateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Body.Slug = channelSlug
	params.Body.Restriction = restriction

	if iteration != nil {
		switch {
		case iteration.ID != "":
			params.Body.IterationID = iteration.ID
		case iteration.Fingerprint != "":
			params.Body.Fingerprint = iteration.Fingerprint
		case iteration.IncrementalVersion > 0:
			params.Body.IncrementalVersion = iteration.IncrementalVersion
		}
	}

	channel, err := client.Packer.PackerServiceCreateChannel(params, nil)
	if err != nil {
		err := err.(*packer_service.PackerServiceCreateChannelDefault)
		return nil, errors.New(err.Payload.Message)
	}

	return channel.GetPayload().Channel, nil
}

// UpdateBucketChannel updates the named channel.
func UpdateBucketChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string, channelSlug string,
	iteration *packermodels.HashicorpCloudPackerIteration, restriction *packermodels.HashicorpCloudPackerUpdateChannelRequestRestriction) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceUpdateChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelSlug
	params.Body.Restriction = restriction

	if iteration != nil {
		switch {
		case iteration.ID != "":
			params.Body.IterationID = iteration.ID
		case iteration.Fingerprint != "":
			params.Body.Fingerprint = iteration.Fingerprint
		case iteration.IncrementalVersion > 0:
			params.Body.IncrementalVersion = iteration.IncrementalVersion
		}
	}

	channel, err := client.Packer.PackerServiceUpdateChannel(params, nil)
	if err != nil {
		err := err.(*packer_service.PackerServiceUpdateChannelDefault)
		return nil, errors.New(err.Payload.Message)
	}

	return channel.GetPayload().Channel, nil
}

// DeleteBucketChannel deletes a channel from the named bucket.
func DeleteBucketChannel(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug, channelSlug string) (*packermodels.HashicorpCloudPackerChannel, error) {
	params := packer_service.NewPackerServiceDeleteChannelParamsWithContext(ctx)
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug
	params.Slug = channelSlug

	req, err := client.Packer.PackerServiceDeleteChannel(params, nil)
	if err != nil {
		err := err.(*packer_service.PackerServiceDeleteChannelDefault)
		return nil, errors.New(err.Payload.Message)
	}

	if !req.IsSuccess() {
		return nil, errors.New("failed to delete channel")
	}

	return nil, nil
}

// ListBucketChannels queries the HCP Packer registry for channels associated to the specified bucket.
func ListBucketChannels(ctx context.Context, client *Client, loc *sharedmodels.HashicorpCloudLocationLocation, bucketSlug string) (*packermodels.HashicorpCloudPackerListChannelsResponse, error) {
	params := packer_service.NewPackerServiceListChannelsParams()
	params.LocationOrganizationID = loc.OrganizationID
	params.LocationProjectID = loc.ProjectID
	params.BucketSlug = bucketSlug

	req, err := client.Packer.PackerServiceListChannels(params, nil)
	if err != nil {
		err := err.(*packer_service.PackerServiceListChannelsDefault)
		return nil, errors.New(err.Payload.Message)
	}

	return req.Payload, nil
}

// handleGetChannelError returns a formatted error for the GetChannel error.
// The upstream API does a good job of providing detailed error messages so we just display the error message, with no status code.
func handleGetChannelError(err *packer_service.PackerServiceGetChannelDefault) error {
	return errors.New(err.Payload.Message)
}

// handleGetIterationError returns a formatted error for the GetIteration error.
// The upstream API does a good job of providing detailed error messages so we just display the error message, with no status code.
func handleGetIterationError(err *packer_service.PackerServiceGetIterationDefault) error {
	return errors.New(err.Payload.Message)
}
